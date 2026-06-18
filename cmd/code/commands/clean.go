package commands

import (
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove any stale repository from staging folder",
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

		err = code.Clean(conf)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
