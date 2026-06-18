package collectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

type RawLsblkData struct {
	Blockdevices []models.Disk `json:"blockdevices"`
}

func getRawDiskWithLsblk() (string, error) {
	command := "lsblk"
	args := []string{"--json"}

	logrus.Debug(
		fmt.Sprintf(
			"Getting list of disks using the following command : %s %s",
			command,
			strings.Join(args, " "),
		),
	)

	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug("Could not run command to list disks")
		return "", executionOutput.ErrorDetails
	}

	return executionOutput.Stdout, nil
}

func processLsblkOutput(lsblkOutput string) ([]models.Disk, error) {
	var rawLsblkData RawLsblkData
	err := json.Unmarshal([]byte(lsblkOutput), &rawLsblkData)
	if err != nil {
		return nil, fmt.Errorf("An error happened while deserializing disks data : %s", err.Error())
	}
	return rawLsblkData.Blockdevices, nil
}

func GetDisks() ([]models.Disk, error) {
	rawData, err := getRawDiskWithLsblk()
	if err != nil {
		return nil, err
	}
	return processLsblkOutput(rawData)
}
