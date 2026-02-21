package commands

import (
	"crypto/tls"
	"fmt"

	"github.com/peeklapp/peekl/pkg/api"
	"github.com/peeklapp/peekl/pkg/bootstrap"
	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
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
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		serverConfig, err := config.NewServerConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		state := bootstrap.GetServerBootstrapState(serverConfig)
		if state == bootstrap.BootstrapNone {
			logrus.Debug("Bootstrap of server was not done, doing it right now.")
			bootstrap.BootstrapServer(serverConfig)
		}

		certsDbEngine, err := certs.NewCertsDatabaseEngine(serverConfig.Certificates.DatabasePath)
		if err != nil {
			logrus.Fatal(err)
		}

		// API Engine
		engine, err := api.NewApiEngine(serverConfig, certsDbEngine)
		if err != nil {
			logrus.Fatal(err)
		}

		// Load certificate for server
		cert, err := tls.LoadX509KeyPair(serverConfig.Certificates.ServerCertificateFilePath, serverConfig.Certificates.ServerCertificateKeyPath)
		if err != nil {
			logrus.Fatal(err)
		}

		// Create TLS Listener
		tlsListener, err := tls.Listen(
			"tcp",
			fmt.Sprintf("%s:%d", serverConfig.Listen.Host, serverConfig.Listen.Port),
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
