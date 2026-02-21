package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
)

func GetServerBootstrapState(serverConfig *config.ServerConfig) BootstrapState {
	bootstrapDoneFileExist := utils.FileExist(serverConfig.Certificates.BootstrapDoneFilePath)
	if bootstrapDoneFileExist {
		return BootstrapComplete
	}
	return BootstrapNone
}

func BootstrapServer(serverConfig *config.ServerConfig) error {
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
	err := certs.CreateCertificateAuthority(
		caParams,
		serverConfig.Certificates.CaCertificateFilePath,
		serverConfig.Certificates.CaCertificateKeyPath,
	)
	if err != nil {
		return err
	}

	err = certs.CreateCertificate(
		serverConfig.Certificates.BootstrapDnsNames,
		serverConfig.Certificates.CaCertificateFilePath,
		serverConfig.Certificates.CaCertificateKeyPath,
		serverConfig.Certificates.ServerCertificateFilePath,
		serverConfig.Certificates.ServerCertificateKeyPath,
	)
	if err != nil {
		return err
	}

	bootstrapDoneFile, err := os.Create(serverConfig.Certificates.BootstrapDoneFilePath)
	if err != nil {
		return err
	}
	defer bootstrapDoneFile.Close()

	return nil
}
