package code

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/peeklapp/peekl/internal/config"
	"github.com/sirupsen/logrus"
)

func Delete(conf *config.CodeConfig, environment string) error {
	logrus.Info(fmt.Sprintf("Starting deletion of env '%s'", environment))

	logrus.Debugf("Deleting folder for environment '%s'", environment)
	environmentFolderPath := filepath.Join(conf.CodeFolder, environment)
	err := os.RemoveAll(environmentFolderPath)
	if err != nil {
		return err
	}
	logrus.Info(fmt.Sprintf("Environment '%s' successfully cleared up", environment))

	return nil
}
