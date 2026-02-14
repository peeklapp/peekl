package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/models"
)

func BootstrapServer(serverConfig *config.ServerConfig, dnsNames []string) error {
	_, err := os.Stat(serverConfig.Certificates.ServerCertificateFilePath)
	if err == nil {
		return fmt.Errorf("The server was apparently already bootstrapped. Make sure to not override any existing certificates.")
	}
	// Make sure any directory that should exist, exist
	dirs := []string{
		serverConfig.Certificates.CaCertificateFilePath,
		serverConfig.Certificates.CaCertificateKeyPath,
		serverConfig.Certificates.ServerCertificateFilePath,
		serverConfig.Certificates.ServerCertificateKeyPath,
		serverConfig.Certificates.DatabasePath,
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

	// Those are straight directories, and not files
	otherDirs := []string{
		serverConfig.Certificates.PendingDirectory,
		serverConfig.Certificates.SignedDirectory,
	}
	for _, dir := range otherDirs {
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return err
			}
		}
	}

	// Create certificate authority
	caParams := models.CertificateAuthorityParameters{
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(10, 0, 0),
	}
	err = certs.CreateCertificateAuthority(
		caParams,
		serverConfig.Certificates.CaCertificateFilePath,
		serverConfig.Certificates.CaCertificateKeyPath,
	)
	if err != nil {
		return err
	}

	err = certs.CreateCertificate(
		dnsNames,
		serverConfig.Certificates.CaCertificateFilePath,
		serverConfig.Certificates.CaCertificateKeyPath,
		serverConfig.Certificates.ServerCertificateFilePath,
		serverConfig.Certificates.ServerCertificateKeyPath,
	)
	if err != nil {
		return err
	}

	return nil
}
