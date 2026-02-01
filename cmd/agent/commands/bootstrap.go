package commands

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
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap agent SSL configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Set log level to verbose
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
		configStruct, err := config.NewAgentConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// If CSR already exist, then host is already bootstrapped
		_, err = os.Stat(configStruct.Certificates.CertificateFilePath)
		if err == nil {
			logrus.Fatal(
				fmt.Errorf("The agent was apparently already bootstrapped. Make sure to not override any existing certificates."),
			)
		}

		// Make sure any directory that should exist, exist
		dirs := []string{
			configStruct.Certificates.CsrFilePath,
			configStruct.Certificates.CaFilePath,
			configStruct.Certificates.CertificateKeyPath,
			configStruct.Certificates.CertificateFilePath,
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

		// Get hostname of the node
		hostname, err := collectors.GetHostname()
		if err != nil {
			logrus.Fatal(err)
		}

		// Create certpool
		certPool := x509.NewCertPool()

		// Create unsecure api client to get CA from server
		bootstrapApiClient, err := client.NewApiClient(*configStruct, true, certPool)
		if err != nil {
			logrus.Fatal(err)
		}

		rootCa, err := bootstrapApiClient.GetRootCA()
		if err != nil {
			logrus.Fatal(err)
		}

		// Write CA file locally
		f, err := os.Create(configStruct.Certificates.CaFilePath)
		if err != nil {
			logrus.Fatal(err)
		}
		defer f.Close()

		_, err = f.Write([]byte(rootCa))
		if err != nil {
			logrus.Fatal(err)
		}

		// Add CA to Client cert pool
		certPool.AppendCertsFromPEM([]byte(rootCa))

		// Create CSR
		err = certs.CreateCertificateSigningRequest(
			hostname,
			configStruct.Certificates.CertificateKeyPath,
			configStruct.Certificates.CsrFilePath,
		)
		if err != nil {
			logrus.Fatal(err)
		}

		// Submit CSR to the server
		csrFile, err := os.ReadFile(configStruct.Certificates.CsrFilePath)
		if err != nil {
			logrus.Fatal(err)
		}
		err = bootstrapApiClient.SubmitCertificateRequest(hostname, string(csrFile))
		if err != nil {
			logrus.Fatal(err)
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
					logrus.Fatal(err)
				}
			}

			// If we've obtain the cert, stop the loop
			if crt != "" {
				// Write certificate locally
				crtFile, err := os.Create(configStruct.Certificates.CertificateFilePath)
				if err != nil {
					logrus.Fatal(err)
				}
				defer crtFile.Close()
				_, err = crtFile.Write([]byte(crt))
				if err != nil {
					logrus.Fatal(err)
				}
				break
			}

			// Sleep for 15 seconds until next retry
			time.Sleep(15 * time.Second)
		}
		logrus.Info("Successfully performed bootstrap against server.")
	},
}
