package bootstrap

import (
	"crypto/x509"
	"os"
	"path/filepath"

	"github.com/peeklapp/peekl/internal/api/client"
	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/facts/collectors"
	"github.com/peeklapp/peekl/internal/utils"
)

func GetAgentBootstrapState(agentConfig *config.AgentConfig) BootstrapState {
	if utils.FileExist(agentConfig.Certificates.BootstrapCompleteFilePath, nil) {
		return BootstrapComplete
	} else if utils.FileExist(agentConfig.Certificates.BootstrapPendingFilePath, nil) {
		return BootstrapPendingCert
	}
	return BootstrapNone
}

func NewBootstrapAgent(agentConfig *config.AgentConfig, token string) error {
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

	// Create CSR
	csr, err := certs.CreateCertificateSigningRequest(
		hostname,
		agentConfig.Certificates.CertificateKeyPath,
		agentConfig.Certificates.CsrFilePath,
	)
	if err != nil {
		return err
	}

	// Enroll agent
	cert, ca, err := bootstrapApiClient.EnrollAgent(csr, token)
	if err != nil {
		return err
	}

	// Write CA file
	caFile, err := os.Create(agentConfig.Certificates.CaFilePath)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(caFile)

	_, err = caFile.Write([]byte(ca))
	if err != nil {
		return err
	}

	// Write cert file
	certFile, err := os.Create(agentConfig.Certificates.CertificateFilePath)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(certFile)

	_, err = certFile.Write([]byte(cert))
	if err != nil {
		return err
	}

	return nil
}
