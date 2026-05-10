package code

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/sirupsen/logrus"
)

func createInfoFile(environment string, codeFolder string, repositoryPath string) error {
	infoFilePath := filepath.Join(codeFolder, fmt.Sprintf("%s.info", environment))
	logrus.Debug(fmt.Sprintf("Creating info file at following path : %s", infoFilePath))

	exist := true
	if _, err := os.Stat(infoFilePath); errors.Is(err, os.ErrNotExist) {
		exist = false
	}

	if exist {
		err := os.Remove(infoFilePath)
		if err != nil {
			return err
		}
	}

	err := os.WriteFile(infoFilePath, []byte(repositoryPath), 0644)
	if err != nil {
		return err
	}

	return nil
}

func Sync(conf *config.CodeConfig, environment string) error {
	if _, err := os.Stat(conf.StagingFolder); errors.Is(err, os.ErrNotExist) {
		logrus.Debug(fmt.Sprintf("%s (staging folder) does not exist, creating it.", conf.StagingFolder))
		if err := os.MkdirAll(conf.StagingFolder, 0750); err != nil {
			return fmt.Errorf("Could not create staging folder : %s", err.Error())
		}
	}

	if _, err := os.Stat(conf.CodeFolder); errors.Is(err, os.ErrNotExist) {
		logrus.Debug(fmt.Sprintf("%s (code folder) does not exist, creating it.", conf.CodeFolder))
		if err := os.MkdirAll(conf.CodeFolder, 0750); err != nil {
			return fmt.Errorf("Could not create code folder : %s", err.Error())
		}
	}

	repositoryName := fmt.Sprintf("%s-%s", environment, uuid.New())
	repositoryPath := filepath.Join(conf.StagingFolder, repositoryName)
	logrus.Debug(fmt.Sprintf("Repository path to clone into is : %s", repositoryPath))

	err := cloneRepository(&conf.Repository, environment, repositoryPath)
	if err != nil {
		logrus.Fatal(err)
	}

	err = createInfoFile(environment, conf.CodeFolder, repositoryPath)
	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}
