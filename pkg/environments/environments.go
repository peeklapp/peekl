package environments

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func EnvironmentNameIsValid(environmentName string) bool {
	r, _ := regexp.Compile("^[A-Za-z0-9_-]+$")
	return r.MatchString(environmentName)
}

func GetEnvironmentFolder(environmentName string, codeDirectory string) (string, error) {
	environmentFolderPath := ""

	infoFilePath := filepath.Join(codeDirectory, fmt.Sprintf("%s.info", environmentName))
	if _, err := os.Stat(infoFilePath); errors.Is(err, os.ErrNotExist) {
		environmentFolderPath = filepath.Join(codeDirectory, environmentName)
	} else {
		data, err := os.ReadFile(infoFilePath)
		if err != nil {
			return environmentFolderPath, err
		}
		environmentFolderPath = string(data)
	}

	return environmentFolderPath, nil
}
