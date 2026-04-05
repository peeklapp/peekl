package collectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type RawLsblkData struct {
	Blockdevices []models.Disk `json:"blockdevices"`
}

func GetDisks() ([]models.Disk, error) {
	var disks []models.Disk

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
		return disks, executionOutput.ErrorDetails
	}

	var rawLsblkData RawLsblkData
	err := json.Unmarshal([]byte(executionOutput.Stdout), &rawLsblkData)
	if err != nil {
		return disks, fmt.Errorf("An error happened while deserializing disks data : %s", err.Error())
	}

	return rawLsblkData.Blockdevices, nil
}
