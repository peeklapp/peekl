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
	runCmd.Flags().BoolP("daemon", "d", false, "Whether to run as daemon or not")
	runCmd.Flags().StringP("environment", "e", "production", "Environment to use")
	runCmd.Flags().StringP("file", "f", "", "File to use (will not try to fetch from the server)")
	runCmd.Flags().StringP("templates", "t", "templates/", "Folder in which to find local templates")
	rootCmd.AddCommand(runCmd)
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

var runCmd = &cobra.Command{
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

		state := bootstrap.GetAgentBootstrapState(agentConfig)
		if err != nil {
			logrus.Fatal(err)
		}

		switch state {
		case bootstrap.BootstrapNone:
			err = bootstrap.BootstrapAgent(agentConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			success, err := bootstrap.TryFetchCertificateFromServer(agentConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			if !success {
				logrus.Fatal("Was not able to fectch certificate from server.")
			}
		case bootstrap.BootstrapPendingCert:
			success, err := bootstrap.TryFetchCertificateFromServer(agentConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			if !success {
				logrus.Fatal("Was not able to fectch certificate from server.")
			}
		}

		apiClient, err := client.NewApiClient(*agentConfig, false, nil)
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
			if !isLocked() {
				createLockfile()
				runAgent(apiClient, agentConfig.Environment)
				deleteLockFile()
			} else {
				logrus.Error("Could not run agent, it's locked. (/tmp/.peekl_run exist)")
			}
		}
	},
}
