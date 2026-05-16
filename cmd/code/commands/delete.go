package commands

import (
	"github.com/peeklapp/peekl/pkg/code"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/environments"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	DeleteCmd.Flags().StringP("environment", "e", "", "Environment to delete from server")
	DeleteCmd.Flags().BoolP("force", "f", false, "Ignore the 'production' branch protection")
	DeleteCmd.MarkFlagRequired("environment")
}

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an enviroment locally",
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

		// Get environment
		environment, err := cmd.Flags().GetString("environment")
		if err != nil {
			logrus.Fatal(err)
		}

		if !environments.EnvironmentNameIsValid(environment) {
			logrus.Fatalf("'%s' is not a valid environment name", environment)
		}

		// Get force
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			logrus.Fatal(err)
		}

		if environment == "production" && !force {
			logrus.Fatal("Not deleting environment 'production'. Use --force if you really want to do it.")
		}

		err = code.Delete(conf, environment)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
