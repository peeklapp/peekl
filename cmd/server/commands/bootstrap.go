package commands

import (
	"github.com/peeklapp/peekl/pkg/bootstrap"
	"github.com/peeklapp/peekl/pkg/config"
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
		serverConfig, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get dns names from command line
		dnsNames, err := cmd.Flags().GetStringArray("names")
		if err != nil {
			logrus.Fatal(err)
		}

		// Bootstrap server
		err = bootstrap.BootstrapServer(serverConfig, dnsNames)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
