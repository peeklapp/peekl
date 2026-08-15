package bootstrap

import (
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
)

func GetServerBootstrapState(serverConfig *config.ServerConfig) BootstrapState {
	bootstrapDoneFileExist := utils.FileExist(serverConfig.Certificates.BootstrapDoneFilePath, nil)
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
	}

	if serverConfig.Database.Type == "sqlite" {
		dirs = append(dirs, serverConfig.Database.Path)
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
