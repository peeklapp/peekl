package bootstrap

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/pkg/api/client"
	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/facts/collectors"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetAgentBootstrapState(agentConfig *config.AgentConfig) BootstrapState {
	if utils.FileExist(agentConfig.Certificates.BootstrapCompleteFilePath) {
		return BootstrapComplete
	} else if utils.FileExist(agentConfig.Certificates.BootstrapPendingFilePath) {
		return BootstrapPendingCert
	}
	return BootstrapNone
}

func BootstrapAgent(agentConfig *config.AgentConfig) error {
	// Make sure any directory that should exist, exist
	dirs := []string{
		agentConfig.Certificates.CsrFilePath,
		agentConfig.Certificates.CaFilePath,
		agentConfig.Certificates.CertificateKeyPath,
		agentConfig.Certificates.CertificateFilePath,
	}
	for _, dir := range dirs {
		basePath := filepath.Dir(dir)
		if _, err := os.Stat(basePath); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(basePath, 0750)
			if err != nil {
				return err
			}
		}
	}

	// Get hostname of the node
	hostname, err := collectors.GetHostname()
	if err != nil {
		return err
	}

	// Create certpool
	certPool := x509.NewCertPool()

	// Create unsecure api client to get CA from server
	bootstrapApiClient, err := client.NewApiClient(*agentConfig, true, certPool)
	if err != nil {
		return err
	}

	rootCa, err := bootstrapApiClient.GetRootCA()
	if err != nil {
		return err
	}

	// Write CA file locally
	caFile, err := os.Create(agentConfig.Certificates.CaFilePath)
	if err != nil {
		return err
	}
	defer caFile.Close()

	_, err = caFile.Write([]byte(rootCa))
	if err != nil {
		return err
	}

	// Add CA to Client cert pool
	certPool.AppendCertsFromPEM([]byte(rootCa))

	// Create CSR
	err = certs.CreateCertificateSigningRequest(
		hostname,
		agentConfig.Certificates.CertificateKeyPath,
		agentConfig.Certificates.CsrFilePath,
	)
	if err != nil {
		return err
	}

	// Submit CSR to the server
	csrFile, err := os.ReadFile(agentConfig.Certificates.CsrFilePath)
	if err != nil {
		return err
	}
	err = bootstrapApiClient.SubmitCertificateRequest(hostname, string(csrFile))
	if err != nil {
		return err
	}

	bootstrapPendingFile, err := os.Create(agentConfig.Certificates.BootstrapPendingFilePath)
	if err != nil {
		return err
	}
	defer bootstrapPendingFile.Close()

	return nil
}

func TryFetchCertificateFromServer(agentConfig *config.AgentConfig) (bool, error) {
	var succes bool

	certPool := x509.NewCertPool()
	caFile, err := os.ReadFile(agentConfig.Certificates.CaFilePath)
	if err != nil {
		return succes, err
	}
	certPool.AppendCertsFromPEM(caFile)

	apiClient, err := client.NewApiClient(*agentConfig, true, certPool)
	if err != nil {
		return succes, err
	}

	hostname, err := collectors.GetHostname()
	if err != nil {
		return succes, err
	}

	csrFile, err := os.ReadFile(agentConfig.Certificates.CsrFilePath)
	if err != nil {
		return succes, err
	}

	for i := 0; i < 5; i++ {
		crt, err := apiClient.RetrieveSignedCertificate(hostname, string(csrFile))
		if err != nil {
			if errors.As(err, &client.HttpError{}) {
				logrus.Error("Certificate not signed yet. You might still have to sign it on the server.")
			} else {
				return succes, err
			}
		}
		if crt != "" {
			crtFile, err := os.Create(agentConfig.Certificates.CertificateFilePath)
			if err != nil {
				return succes, err
			}
			defer crtFile.Close()
			_, err = crtFile.Write([]byte(crt))
			if err != nil {
				return succes, err
			}
			break
		}
		time.Sleep(15 * time.Second)
	}

	if utils.FileExist(agentConfig.Certificates.CertificateFilePath) {
		logrus.Info("Successfully performed bootstrap against server.")

		bootstrapDoneFile, err := os.Create(agentConfig.Certificates.BootstrapCompleteFilePath)
		if err != nil {
			return succes, err
		}
		defer bootstrapDoneFile.Close()

		succes = true
	}

	return succes, nil
}
