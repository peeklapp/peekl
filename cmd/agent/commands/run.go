package commands

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redat00/peekl/pkg/catalog"
	"github.com/redat00/peekl/pkg/facts"
	"github.com/redat00/peekl/pkg/models"

	"github.com/redat00/peekl/pkg/api/client"
	"github.com/redat00/peekl/pkg/config"

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

func runAgent(client *client.Client) {
	// Create rawCatalog
	var rawCatalog models.RawCatalog

	// Collect facts
	var err error
	facter := facts.NewFacter()
	rawCatalog.Facts, err = facter.Collect()
	if err != nil {
		logrus.Fatal(err)
	}

	rawCatalog.GlobalResources, rawCatalog.Roles, rawCatalog.Tags, rawCatalog.Variables, err = client.GetCatalog()
	if err != nil {
		logrus.Error(err)
		return
	}

	catalog, err := catalog.NewCatalog(rawCatalog)
	if err != nil {
		logrus.Error(err)
		return
	}

	err = catalog.Process()
	if err != nil {
		logrus.Error(err)
		return
	}
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

		// Get configuration
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			logrus.Fatal(err)
		}
		configStruct, err := config.NewAgentConfiguration(configPath)
		if err != nil {
			logrus.Fatal(err)
		}

		client, err := client.NewApiClient(*configStruct, false, nil)
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
					runAgent(client)
					deleteLockFile()
				} else {
					logrus.Error("Could not run agent, it's lock. (/tmp/.peekl_run exist)")
				}
				logrus.Info(fmt.Sprintf("Next run in %d seconds.", configStruct.Daemon.LoopTime))
				time.Sleep(time.Duration(configStruct.Daemon.LoopTime) * time.Second)
			}
		} else {
			if !isLocked() {
				createLockfile()
				runAgent(client)
				deleteLockFile()
			} else {
				logrus.Error("Could not run agent, it's lock. (/tmp/.peekl_run exist)")
			}
		}
	},
}
