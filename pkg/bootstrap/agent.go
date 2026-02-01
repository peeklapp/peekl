package bootstrap

import (
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/redat00/peekl/pkg/api/client"
	"github.com/redat00/peekl/pkg/certs"
	"github.com/redat00/peekl/pkg/config"
	"github.com/redat00/peekl/pkg/facts/collectors"
	"github.com/sirupsen/logrus"
)

func BootstrapAgent(agentConfig *config.AgentConfig) error {
	// If CSR already exist, then host is already bootstrapped
	_, err := os.Stat(agentConfig.Certificates.CertificateFilePath)
	if err == nil {
		return fmt.Errorf("The agent was apparently already bootstrapped. Make sure to not override any existing certificates.")
	}

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
	f, err := os.Create(agentConfig.Certificates.CaFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(rootCa))
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

	// Try to get the certificate from the server every n seconds
	for i := 0; i < 5; i++ {
		// Try to get the certificate
		crt, err := bootstrapApiClient.RetrieveSignedCertificate(hostname, string(csrFile))

		// Check if we've had an error
		if err != nil {
			if errors.As(err, &client.HttpError{}) {
				logrus.Error("Certificate not signed yet. You might still have to sign in on the server.")
			} else {
				return err
			}
		}

		// If we've obtain the cert, stop the loop
		if crt != "" {
			// Write certificate locally
			crtFile, err := os.Create(agentConfig.Certificates.CertificateFilePath)
			if err != nil {
				return err
			}
			defer crtFile.Close()
			_, err = crtFile.Write([]byte(crt))
			if err != nil {
				return err
			}
			break
		}

		// Sleep for 15 seconds until next retry
		time.Sleep(15 * time.Second)
	}
	logrus.Info("Successfully performed bootstrap against server.")

	return nil
}
