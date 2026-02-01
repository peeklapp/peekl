package commands

import (
	"crypto/tls"
	"fmt"

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

		// TODO: IMPLEMENT CHECK IF BOOTSTRAP WAS DONE, DO IT IF NOT

		// Load configuration
		configStruct, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
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

		// Load certificate for server
		cert, err := tls.LoadX509KeyPair(configStruct.Certificates.ServerCertificateFilePath, configStruct.Certificates.ServerCertificateKeyPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create TLS Listener
		tlsListener, err := tls.Listen(
			"tcp",
			fmt.Sprintf("%s:%d", configStruct.Listen.Host, configStruct.Listen.Port),
			&tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequestClientCert,
				MinVersion:   tls.VersionTLS12,
			},
		)
		if err != nil {
			logrus.Fatal(err)
		}

		// Start server
		engine.Listener(tlsListener)
	},
}
