package code

import (
	"os"
	"path/filepath"
	"strconv"
)

const (
	TarballExtension = ".tar.zst"
	CodeTarballName  = "code" + TarballExtension
)

func GetHighestIdInEnvironment(environmentFolder string) (int, error) {
	entries, err := os.ReadDir(environmentFolder)
	if err != nil {
		return 0, err
	}

	var maxValue int = -1
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n, err := strconv.Atoi(e.Name())
		if err != nil {
			return 0, err
		}
		if n > maxValue {
			maxValue = n
		}
	}

	if maxValue < 0 {
		return 0, nil
	}

	return maxValue, nil
}

func GetLatestVersionInEnvironment(root *os.Root, environmentFolder string) (string, error) {
	latestFile, err := root.ReadFile(filepath.Join(environmentFolder, "latest"))
	if err != nil {
		return "", err
	}
	return string(latestFile), nil
}
