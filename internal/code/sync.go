package code

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/peeklapp/peekl/internal/config"
	"github.com/sirupsen/logrus"
)

func createLatestFile(environment string, codeFolder string, latestId int) error {
	latestFilePath := filepath.Join(codeFolder, environment, "latest")

	if _, err := os.Stat(latestFilePath); !errors.Is(err, os.ErrNotExist) {
		err := os.Remove(latestFilePath)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(latestFilePath, []byte(strconv.Itoa(latestId)), 0644)
	if err != nil {
		return err
	}

	return nil
}

func Sync(conf *config.CodeConfig, environment string) error {
	if _, err := os.Stat(filepath.Join(conf.CodeFolder, environment)); errors.Is(err, os.ErrNotExist) {
		logrus.Debug(fmt.Sprintf("%s (code folder) does not exist, creating it.", filepath.Join(conf.CodeFolder, environment)))
		if err := os.MkdirAll(filepath.Join(conf.CodeFolder, environment), 0750); err != nil {
			return fmt.Errorf("Could not create code folder : %s", err.Error())
		}
	}

	logrus.Debugf("Creating temporary directory to clone into for environment '%s'", environment)
	tempDir, err := os.MkdirTemp("", "peekl")
	if err != nil {
		return fmt.Errorf("Could not create temporary directory : %s", err.Error())
	}
	defer os.RemoveAll(tempDir)

	logrus.Debugf("Cloning repository for environment '%s'", environment)
	err = cloneRepository(&conf.Repository, environment, tempDir)
	if err != nil {
		return fmt.Errorf("Could not clone repository : %s", err.Error())
	}

	logrus.Debugf("Finding highest ID to use for environment '%s'", environment)
	id, err := GetHighestIdInEnvironment(filepath.Join(conf.CodeFolder, environment))
	if err != nil {
		return fmt.Errorf("Could not get highest ID : %s", err.Error())
	}
	id = id + 1

	logrus.Debugf("Generating global code archive for environment '%s'", environment)
	err = GenerateCodeArchive(tempDir, filepath.Join(conf.CodeFolder, environment, strconv.Itoa(id)))
	if err != nil {
		return fmt.Errorf("Could not generate code archive : %s", err.Error())
	}

	logrus.Debugf("Generating nodes archives for environment '%s'", environment)
	err = GenerateNodesArchives(tempDir, filepath.Join(conf.CodeFolder, environment, strconv.Itoa(id), "nodes"))
	if err != nil {
		return fmt.Errorf("Could not generate nodes archives : %s", err.Error())
	}

	logrus.Debugf("Creating the latest file for environment '%s'", environment)
	err = createLatestFile(environment, conf.CodeFolder, id)
	if err != nil {
		return fmt.Errorf("Could not create the latest file for environment : %s", err.Error())
	}

	return nil
}
