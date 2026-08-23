package commands

import (
	"github.com/peeklapp/peekl/internal/bootstrap"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var EnrollCmd = &cobra.Command{
	Use:   "enroll [token]",
	Args:  cobra.ExactArgs(1),
	Short: "Enroll the agent with the Peekl server",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		agentConfig, err := config.NewAgentConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		err = bootstrap.NewBootstrapAgent(agentConfig, args[0])
		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Info("Successfully performed enrollment against server")
	},
}
