package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/peeklapp/peekl/internal/bootstrap"
	"github.com/peeklapp/peekl/internal/catalog"
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/environments"
	"github.com/peeklapp/peekl/internal/facts"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"

	"github.com/peeklapp/peekl/internal/api/client"
	"github.com/peeklapp/peekl/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	RunCmd.Flags().BoolP("daemon", "d", false, "Whether to run as daemon or not")
	RunCmd.Flags().StringP("environment", "e", "production", "Environment to use")
}

func isLocked() bool {
	return utils.FileExist("/tmp/.peekl_run", nil)
}

func createLockfile() error {
	if _, err := os.Create("/tmp/.peekl_run"); err != nil {
		return err
	}
	return nil
}

func deleteLockFile() error {
	return os.Remove("/tmp/.peekl_run")
}

// Verify that a cache is still valid
func isCacheValid(filePath string, expectedHash string) (bool, error) {
	if !utils.FileExist(filePath, nil) {
		return false, nil
	}

	checksum, err := utils.GetMd5CheckumForFile(filePath, nil)
	if err != nil {
		return false, err
	}

	return expectedHash == checksum, nil
}

func deleteExtractDir(extractDirPath string) {
	err := os.RemoveAll(extractDirPath)
	if err != nil {
		logrus.Fatal(err)
	}
}

func runAgent(client *client.Client, environment string, cachePath string) {
	var rawCatalog models.RawCatalog
	rawCatalog.Environment = environment
	rawCatalog.ApiClient = client

	var err error
	facter := facts.NewFacter()
	logrus.Debug("Collecting facts")
	rawCatalog.Facts, err = facter.Collect()
	if err != nil {
		logrus.Fatal(err)
	}

	// Getting what to download
	logrus.Debug("Getting URL path for tarballs")
	nodeTarballUrl, nodeTarballHash, codeTarballUrl, codeTarballHash, err := client.GetCatalog(environment)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Debug("Creating global cache folder if it doesn't exist")
	if !utils.FileExist(cachePath, nil) {
		if err := os.MkdirAll(cachePath, 0750); err != nil {
			logrus.Fatal(err)
		}
	}

	logrus.Debug("Creating environment cache folder if it doesn't exist")
	if !utils.FileExist(filepath.Join(cachePath, environment), nil) {
		if err := os.Mkdir(filepath.Join(cachePath, environment), 0750); err != nil {
			logrus.Fatal(err)
		}
	}

	logrus.Debug("Checking local cache to see if we need to download files again")
	expectedNodeFilePath := filepath.Join(cachePath, environment, "node"+code.TarballExtension)
	nodeFileCacheValid, err := isCacheValid(expectedNodeFilePath, nodeTarballHash)
	if err != nil {
		logrus.Fatal(err)
	}
	expectedCodeFilePath := filepath.Join(cachePath, environment, code.CodeTarballName)
	codeFileCacheValid, err := isCacheValid(expectedCodeFilePath, codeTarballHash)
	if err != nil {
		logrus.Fatal(err)
	}

	if !nodeFileCacheValid {
		logrus.Debugf("Local node cache is not valid, downloading with path : %s", nodeTarballUrl)
		err = client.DownloadFile(nodeTarballUrl, expectedNodeFilePath)
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Debug("Local node cache is valid, not downloading.")
	}

	if !codeFileCacheValid {
		logrus.Debugf("Local code cache is not valid, downloading with path : %s", codeTarballUrl)
		err = client.DownloadFile(codeTarballUrl, expectedCodeFilePath)
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Debug("Local code cache is valid, not downloading.")
	}

	logrus.Debug("Creating tarball temporary extraction directory")
	extractDir, err := os.MkdirTemp("", "peekl")
	if err != nil {
		logrus.Fatal(err)
	}
	defer deleteExtractDir(extractDir)

	logrus.Debug("Extracting the archives")
	archives := []string{expectedCodeFilePath, expectedNodeFilePath}
	for _, arch := range archives {
		logrus.Debugf("Extracting first archive at path '%s' into '%s'", arch, extractDir)
		err := code.DecompressArchive(arch, extractDir)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	logrus.Debug("Compiling raw catalog based on extracted archives")
	rawCatalog.GlobalResources, rawCatalog.Roles, rawCatalog.Tags, rawCatalog.Variables, err = catalog.CompileCatalog(extractDir, rawCatalog.Facts.Hostname)
	if err != nil {
		logrus.Fatal(err)
	}
	rawCatalog.CodePath = extractDir

	logrus.Debug("Creating catalog from raw catalog")
	catalog, err := catalog.NewCatalog(rawCatalog)
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Debug("Validating the catalog")
	valid, err := catalog.Validate()
	if err != nil {
		logrus.Fatal(err)
	}

	if valid {
		logrus.Info("Catalog is valid, running")
		err = catalog.Process()
		if err != nil {
			logrus.Fatal(err)
		}
	} else {
		logrus.Fatal("Catalog is not valid. Not running.")
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
			return fmt.Errorf("could not fetch certificate from server")
		}
	case bootstrap.BootstrapPendingCert:
		success, err := bootstrap.TryFetchCertificateFromServer(config)
		if err != nil {
			return err
		}
		if !success {
			return fmt.Errorf("could not fetch certificate from server")
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

		// Get environment to use from CLI
		environment := ""
		if cmd.Flags().Changed("environment") {
			environment, err = cmd.Flags().GetString("environment")
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			if agentConfig.Environment != environment {
				environment = agentConfig.Environment
			}
		}

		// Validate that the environment is valid
		if !environments.EnvironmentNameIsValid(environment) {
			logrus.Fatalf("'%s' is not a valid environment name", environment)
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
					if err := createLockfile(); err != nil {
						logrus.Fatalf("Could not create lock due to the following error : %s", err.Error())
					}
					runAgent(apiClient, environment, agentConfig.Caching.Path)
					if err := deleteLockFile(); err != nil {
						logrus.Fatalf("Could not delete lock file due to the following error : %s", err.Error())
					}
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
				err = createLockfile()
				if err != nil {
					logrus.Fatalf("Could not create lock due to the following error : %s", err.Error())
				}
				runAgent(apiClient, environment, agentConfig.Caching.Path)
				if err := deleteLockFile(); err != nil {
					logrus.Fatalf("Could not delete lock file due to the following error : %s", err.Error())
				}
			} else {
				logrus.Error("Could not run agent, it's locked. (/tmp/.peekl_run exist)")
			}
		}
	},
}
