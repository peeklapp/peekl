package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/redat00/peekl/pkg/certs"
	"github.com/redat00/peekl/pkg/config"
	"github.com/redat00/peekl/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().StringArray("names", []string{"peekl"}, "List of DNS names to set in server certificate")
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap the server by creating the certificate authority",
	Run: func(cmd *cobra.Command, args []string) {
		// Get verbosity
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Get configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		_, err = os.Stat(configStruct.Certificates.ServerCertificateFilePath)
		if err == nil {
			logrus.Fatal(
				fmt.Errorf("The server was apparently already bootstrapped. Make sure to not override any existing certificates."),
			)
		}

		// Make sure any directory that should exist, exist
		dirs := []string{
			configStruct.Certificates.CaCertificateFilePath,
			configStruct.Certificates.CaCertificateKeyPath,
			configStruct.Certificates.ServerCertificateFilePath,
			configStruct.Certificates.ServerCertificateKeyPath,
			configStruct.Certificates.DatabasePath,
		}
		for _, dir := range dirs {
			basePath := filepath.Dir(dir)
			if _, err := os.Stat(basePath); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(basePath, 0750)
				if err != nil {
					logrus.Fatal(err)
				}
			}
		}

		// Those are straight directories, and not files
		otherDirs := []string{
			configStruct.Certificates.PendingDirectory,
			configStruct.Certificates.SignedDirectory,
		}
		for _, dir := range otherDirs {
			if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(dir, 0750)
				if err != nil {
					logrus.Fatal(err)
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
			configStruct.Certificates.CaCertificateFilePath,
			configStruct.Certificates.CaCertificateKeyPath,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		dnsNames, err := cmd.Flags().GetStringArray("names")
		if err != nil {
			logrus.Fatal(err)
		}

		err = certs.CreateCertificate(
			dnsNames,
			configStruct.Certificates.CaCertificateFilePath,
			configStruct.Certificates.CaCertificateKeyPath,
			configStruct.Certificates.ServerCertificateFilePath,
			configStruct.Certificates.ServerCertificateKeyPath,
		)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
