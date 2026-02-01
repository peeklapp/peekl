package commands

import (
	"github.com/redat00/peekl/pkg/bootstrap"
	"github.com/redat00/peekl/pkg/config"
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
		agentConfig, err := config.NewAgentConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		err = bootstrap.BootstrapAgent(agentConfig)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
