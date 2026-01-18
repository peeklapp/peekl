package commands

import (
	"errors"
	"fmt"
	"os"
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

		// Create default directories for certificates
		_, err = os.Stat(configStruct.Certificates.CaDirectory)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				os.MkdirAll(configStruct.Certificates.CaDirectory, 0700)
			}
		} else {
			logrus.Fatal(
				fmt.Errorf(
					"The directory for certificates apparently already exist, you should make sure to not overwrite a pre-existing bootstrap of Peekl",
				),
			)
		}

		// Create certificate authority
		caParams := models.CertificateAuthorityParameters{
			NotBefore: time.Now(),
			NotAfter:  time.Now().AddDate(10, 0, 0),
		}
		err = certs.CreateCertificateAuthority(configStruct.Certificates.CaDirectory, caParams)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create server certificate
		_, err = os.Stat(configStruct.Certificates.ServerDirectory)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				os.MkdirAll(configStruct.Certificates.ServerDirectory, 0700)
			} else {
				logrus.Fatal(err)
			}
		}

		dnsNames, err := cmd.Flags().GetStringArray("names")
		if err != nil {
			logrus.Fatal(err)
		}

		err = certs.CreateCertificate(configStruct.Certificates.CaDirectory, configStruct.Certificates.ServerDirectory, "server", dnsNames)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
