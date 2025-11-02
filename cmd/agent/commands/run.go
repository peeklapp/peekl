package commands

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/redat00/peekl/pkg/catalog"
	"github.com/redat00/peekl/pkg/facts"
	"github.com/redat00/peekl/pkg/models"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().StringP("environment", "e", "production", "Environment to use")
	runCmd.Flags().StringP("file", "f", "", "File to use (will not try to fetch from the server)")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the agent",
	Run: func(cmd *cobra.Command, args []string) {
		// Set log level to debug
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			logrus.Fatal(err)
		}

		if verbose {
			logrus.SetLevel(logrus.DebugLevel)
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			logrus.Fatal(err)
		}

		// Collect facts
		facter := facts.NewFacter()
		facts, err := facter.Collect()
		if err != nil {
			logrus.Fatal(err)
		}

		var resources []models.Resource

		if file != "" {
			logrus.Debug("Local file provided, reading resources from file")
			f, err := os.ReadFile(file)
			if err != nil {
				logrus.Fatal(err)
			}

			err = yaml.Unmarshal(f, &resources)
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			logrus.Debug("No local file provided, contacting server to get ressources")
		}

		catalog, err := catalog.NewCatalog(resources, facts)
		if err != nil {
			logrus.Fatal(err)
		}

		err = catalog.Process()
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
