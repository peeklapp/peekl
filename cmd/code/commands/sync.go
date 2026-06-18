package commands

import (
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/environments"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	SyncCmd.Flags().StringP("environment", "e", "production", "Environment to sync from distant repository")
}

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync an enviroment from remote repository",
	Run: func(cmd *cobra.Command, args []string) {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}
		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Load configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		conf, err := config.NewCodeConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Get environment to sync
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			logrus.Fatal(err)
		}

		if !environments.EnvironmentNameIsValid(environment) {
			logrus.Fatalf("'%s' is not a valid environment name", environment)
		}

		// Sync environment
		if err := code.Sync(conf, environment); err != nil {
			logrus.Fatal(err)
		}
	},
}
