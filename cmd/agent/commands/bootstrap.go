package commands

import (
	"errors"
	"fmt"
	"os"

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

		// Create default directories for certificates
		_, err = os.Stat(configStruct.Certificates.Directory)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				os.MkdirAll(configStruct.Certificates.Directory, 0700)
			} else {
				logrus.Fatal(
					fmt.Errorf(
						"The directory for certificates apparently already exist, you should make sure to not overwrite a pre-existing bootstrap of the Peekl agent",
					),
				)
			}
		}

		// Get hostname
		hostname, err := collectors.GetHostname()
		if err != nil {
			logrus.Fatal(err)
		}

		// Create CSR
		err = certs.CreateCertificateSigningRequest(configStruct.Certificates.Directory, hostname)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
