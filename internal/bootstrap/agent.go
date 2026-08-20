package bootstrap

import (
	"crypto/x509"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/internal/api/client"
	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/facts/collectors"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

func GetAgentBootstrapState(agentConfig *config.AgentConfig) BootstrapState {
	if utils.FileExist(agentConfig.Certificates.BootstrapCompleteFilePath, nil) {
		return BootstrapComplete
	} else if utils.FileExist(agentConfig.Certificates.BootstrapPendingFilePath, nil) {
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
		if !utils.FileExist(basePath, nil) {
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
	defer utils.CloseWithoutError(caFile)

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
	defer utils.CloseWithoutError(bootstrapPendingFile)

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

	csrFile, err := os.ReadFile(agentConfig.Certificates.CsrFilePath)
	if err != nil {
		return succes, err
	}

	signature := certs.GetCertificateSigningRequestSignature(string(csrFile))

	for i := 0; i < 5; i++ {
		crt, err := apiClient.RetrieveSignedCertificate(signature)
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
			defer utils.CloseWithoutError(crtFile)
			_, err = crtFile.Write([]byte(crt))
			if err != nil {
				return succes, err
			}
			break
		}
		time.Sleep(15 * time.Second)
	}

	if utils.FileExist(agentConfig.Certificates.CertificateFilePath, nil) {
		logrus.Info("Successfully performed bootstrap against server.")

		bootstrapDoneFile, err := os.Create(agentConfig.Certificates.BootstrapCompleteFilePath)
		if err != nil {
			return succes, err
		}
		defer utils.CloseWithoutError(bootstrapDoneFile)

		succes = true
	}

	return succes, nil
}
