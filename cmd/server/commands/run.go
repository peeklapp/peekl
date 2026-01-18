package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/redat00/peekl/pkg/api"
	"github.com/redat00/peekl/pkg/certs"
	"github.com/redat00/peekl/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the server",
	Run: func(cmd *cobra.Command, args []string) {
		// Get configuration file path
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}

		// Load configuration
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create directories if they don't exist
		dirs := []string{
			configStruct.Certificates.MainDirectory,
			configStruct.Certificates.CaDirectory,
			configStruct.Certificates.PendingDirectory,
			configStruct.Certificates.SignedDirectory,
		}
		for _, dir := range dirs {
			if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
				err := os.MkdirAll(dir, 0750)
				if err != nil {
					logrus.Fatal(err)
				}
			}
		}

		certsDbEngine, err := certs.NewCertsDatabaseEngine(configStruct.Certificates.DatabasePath)
		if err != nil {
			logrus.Fatal(err)
		}

		// API Engine
		engine, err := api.NewApiEngine(configStruct, certsDbEngine)
		if err != nil {
			logrus.Fatal(err)
		}

		// Start server
		err = engine.ListenTLS(
			fmt.Sprintf("%s:%d", configStruct.Listen.Host, configStruct.Listen.Port),
			fmt.Sprintf("%s/server.crt", configStruct.Certificates.ServerDirectory),
			fmt.Sprintf("%s/server.key", configStruct.Certificates.ServerDirectory),
		)
		if err != nil {
			logrus.Fatal(err)
		}
	},
}
