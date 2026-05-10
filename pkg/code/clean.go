package code

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peeklapp/peekl/pkg/config"
	"github.com/sirupsen/logrus"
)

func getAllDirsInStaging(stagingFolder string) ([]string, error) {
	entries, err := os.ReadDir(stagingFolder)
	if err != nil {
		return nil, err
	}

	var dirsInStaging []string
	for _, e := range entries {
		if e.IsDir() {
			dirsInStaging = append(dirsInStaging, filepath.Join(stagingFolder, e.Name()))
		}
	}

	return dirsInStaging, nil
}

func getAllReferencedDirsInCode(codeFolder string) ([]string, error) {
	entries, err := os.ReadDir(codeFolder)
	if err != nil {
		return nil, err
	}

	var referenceDirs []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".info") {
			data, err := os.ReadFile(filepath.Join(codeFolder, e.Name()))
			if err != nil {
				return nil, err
			}
			referenceDirs = append(referenceDirs, string(data))
		}
	}

	return referenceDirs, nil
}

func findDirsToDelete(dirsInStaging []string, referencedDirs []string) []string {
	var dirsToDelete []string

	for _, dir := range dirsInStaging {
		found := false
		for _, ref := range referencedDirs {
			if dir == ref {
				found = true
			}
		}
		if !found {
			dirsToDelete = append(dirsToDelete, dir)
		}
	}

	return dirsToDelete
}

func Clean(conf *config.CodeConfig) error {
	logrus.Info("Getting all folders in staging")
	dirsInStaging, err := getAllDirsInStaging(conf.StagingFolder)
	if err != nil {
		return err
	}

	logrus.Info("Getting all referenced folders")
	referencedDirs, err := getAllReferencedDirsInCode(conf.CodeFolder)
	if err != nil {
		return err
	}

	logrus.Info("Filtering folders not referenced")
	dirsToDelete := findDirsToDelete(dirsInStaging, referencedDirs)
	logrus.Info(fmt.Sprintf("Found %d stale folder to delete.", len(dirsToDelete)))

	if len(dirsToDelete) > 0 {
		logrus.Info("Deleting folders")
		for _, dir := range dirsToDelete {
			err := os.RemoveAll(dir)
			if err != nil {
				return err
			}

		}
	}

	return nil
}
