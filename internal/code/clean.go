package code

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/peeklapp/peekl/internal/config"
	"github.com/sirupsen/logrus"
)

type DirInCodeFolder struct {
	Path string
	Id   int
	Name string
}

func getListOfEnvironments(codeFolder string) ([]string, error) {
	entries, err := os.ReadDir(codeFolder)
	if err != nil {
		return nil, err
	}

	var environmentsDir []string
	for _, e := range entries {
		if e.IsDir() {
			environmentsDir = append(environmentsDir, e.Name())
		}
	}

	return environmentsDir, nil
}

func getDirsInCodeFolder(environmentPath string) ([]DirInCodeFolder, error) {
	entries, err := os.ReadDir(environmentPath)
	if err != nil {
		return nil, err
	}

	var dirsInCodeFolder []DirInCodeFolder
	for _, e := range entries {
		if e.IsDir() {
			dirId, err := strconv.Atoi(e.Name())
			if err != nil {
				return dirsInCodeFolder, err
			}
			dirsInCodeFolder = append(
				dirsInCodeFolder,
				DirInCodeFolder{
					Path: filepath.Join(environmentPath, e.Name()),
					Name: e.Name(),
					Id:   dirId,
				},
			)
		}
	}

	return dirsInCodeFolder, nil
}

func findDirsToDelete(environmentPath string, toKeep int) ([]DirInCodeFolder, error) {
	dirsInCodeFolder, err := getDirsInCodeFolder(environmentPath)
	if err != nil {
		return nil, err
	}

	highestId, err := GetHighestIdInEnvironment(environmentPath)
	if err != nil {
		return nil, err
	}

	lowestAcceptableId := highestId - toKeep
	if lowestAcceptableId <= 0 {
		lowestAcceptableId = 0
	}

	var dirsToDelete []DirInCodeFolder
	for _, dirInCodeFolder := range dirsInCodeFolder {
		if dirInCodeFolder.Id <= lowestAcceptableId {
			dirsToDelete = append(dirsToDelete, dirInCodeFolder)
		}
	}

	return dirsToDelete, nil
}

func Clean(conf *config.CodeConfig) error {
	logrus.Info("Finding list of existing environments")
	existingEnvironments, err := getListOfEnvironments(conf.CodeFolder)
	if err != nil {
		return fmt.Errorf("Could not obtain list of environments : %s", err.Error())
	}

	for _, env := range existingEnvironments {
		environmentPath := filepath.Join(conf.CodeFolder, env)
		logrus.Infof("Finding list of folders to be deleted for environment '%s'", env)
		dirsToDelete, err := findDirsToDelete(environmentPath, conf.Keep)
		if err != nil {
			return fmt.Errorf("Could not obtain the list of folders to delete : %s", err.Error())
		}
		logrus.Info(fmt.Sprintf("Found %d stale folder to delete.", len(dirsToDelete)))

		if len(dirsToDelete) > 0 {
			logrus.Info("Deleting folders")
			for _, dir := range dirsToDelete {
				err := os.RemoveAll(dir.Path)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
