package commands

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/peeklapp/peekl/pkg/bootstrap"
	"github.com/peeklapp/peekl/pkg/catalog"
	"github.com/peeklapp/peekl/pkg/facts"
	"github.com/peeklapp/peekl/pkg/models"

	"github.com/peeklapp/peekl/pkg/api/client"
	"github.com/peeklapp/peekl/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RunCmd.Flags().BoolP("daemon", "d", false, "Whether to run as daemon or not")
	RunCmd.Flags().StringP("environment", "e", "production", "Environment to use")
	RunCmd.Flags().StringP("file", "f", "", "File to use (will not try to fetch from the server)")
	RunCmd.Flags().StringP("templates", "t", "templates/", "Folder in which to find local templates")
}

func isLocked() bool {
	if _, err := os.Stat("/tmp/.peekl_run"); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func createLockfile() {
	os.Create("/tmp/.peekl_run")
}

func deleteLockFile() {
	os.Remove("/tmp/.peekl_run")
}

func runAgent(client *client.Client, environment string) {
	// Create rawCatalog
	var rawCatalog models.RawCatalog

	// Collect facts
	var err error
	facter := facts.NewFacter()
	rawCatalog.Facts, err = facter.Collect()
	if err != nil {
		logrus.Fatal(err)
	}

	rawCatalog.GlobalResources, rawCatalog.Roles, rawCatalog.Tags, rawCatalog.Variables, err = client.GetCatalog(environment)
	if err != nil {
		logrus.Error(err)
		return
	}

	catalog, err := catalog.NewCatalog(rawCatalog)
	if err != nil {
		logrus.Error(err)
		return
	}

	valid, err := catalog.Validate()
	if err != nil {
		logrus.Error(err)
	}

	if valid {
		logrus.Info("Catalog is valid, running")
		err = catalog.Process()
		if err != nil {
			logrus.Error(err)
			return
		}
	} else {
		logrus.Error("Catalog is not valid. Not running.")
	}
}

func performBootstrap(config *config.AgentConfig) error {
	state := bootstrap.GetAgentBootstrapState(config)

	switch state {
	case bootstrap.BootstrapNone:
		err := bootstrap.BootstrapAgent(config)
		if err != nil {
			return err
		}
		success, err := bootstrap.TryFetchCertificateFromServer(config)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("Could not fetch certificate from server")
		}
	case bootstrap.BootstrapPendingCert:
		success, err := bootstrap.TryFetchCertificateFromServer(config)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("Could not fetch certificate from server")
		}
	}

	return nil
}

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the agent",
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

		// Get if should be run as daemon
		daemon, err := cmd.Flags().GetBool("daemon")
		if err != nil {
			logrus.Fatal(err)
		}

		if daemon {
			for {
				err = performBootstrap(agentConfig)
				if err != nil {
					logrus.Error(err)
				} else {
					break
				}
				logrus.Info("Retrying in 60 seconds")
				time.Sleep(time.Duration(60) * time.Second)
			}

			apiClient, err := client.NewApiClient(*agentConfig, false, nil)
			if err != nil {
				logrus.Fatal(err)
			}

			for {
				if !isLocked() {
					createLockfile()
					runAgent(apiClient, agentConfig.Environment)
					deleteLockFile()
				} else {
					logrus.Error("Could not run agent, it's locked. (/tmp/.peekl_run exist)")
				}
				logrus.Info(fmt.Sprintf("Next run in %d seconds.", agentConfig.Daemon.LoopTime))
				time.Sleep(time.Duration(agentConfig.Daemon.LoopTime) * time.Second)
			}
		} else {
			err = performBootstrap(agentConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			if !isLocked() {
				apiClient, err := client.NewApiClient(*agentConfig, false, nil)
				if err != nil {
					logrus.Fatal(err)
				}
				createLockfile()
				runAgent(apiClient, agentConfig.Environment)
				deleteLockFile()
			} else {
				logrus.Error("Could not run agent, it's locked. (/tmp/.peekl_run exist)")
			}
		}
	},
}
