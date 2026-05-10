package code

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peeklapp/peekl/pkg/config"
	"github.com/sirupsen/logrus"
)

func getAllStagedRepositoryForEnv(stagingFolder string, environment string) ([]string, error) {
	entries, err := os.ReadDir(stagingFolder)
	if err != nil {
		return nil, err
	}

	var stagedRepositoryForEnv []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), environment) {
			stagedRepositoryForEnv = append(stagedRepositoryForEnv, filepath.Join(stagingFolder, e.Name()))
		}
	}

	return stagedRepositoryForEnv, nil
}

func Delete(conf *config.CodeConfig, environment string) error {
	logrus.Info(fmt.Sprintf("Starting deletion of env '%s'", environment))

	infoFilePath := filepath.Join(conf.CodeFolder, fmt.Sprintf("%s.info", environment))

	logrus.Debug(fmt.Sprintf("Checking if info file for env '%s' exist", environment))
	infoFileExist := true
	if _, err := os.Stat(infoFilePath); errors.Is(err, os.ErrNotExist) {
		infoFileExist = false
	}

	if infoFileExist {
		logrus.Debug(fmt.Sprintf("Info file for env '%s' exist, deleting it", environment))
		err := os.Remove(infoFilePath)
		if err != nil {
			return err
		}
	}

	stagedRepositoryForEnv, err := getAllStagedRepositoryForEnv(conf.StagingFolder, environment)
	if err != nil {
		return err
	}

	if len(stagedRepositoryForEnv) > 0 {
		logrus.Debug(fmt.Sprintf("Found %d staged folders related to env '%s', deleting them", len(stagedRepositoryForEnv), environment))
		for _, dir := range stagedRepositoryForEnv {
			logrus.Debug(dir)
			err := os.RemoveAll(dir)
			if err != nil {
				return err
			}
		}
	}

	logrus.Info(fmt.Sprintf("Environment '%s' successfully cleared up", environment))

	return nil
}
