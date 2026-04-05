package collectors

import (
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetLsbData() (models.LsbData, error) {
	var lsbData models.LsbData

	command := "lsb_release"
	args := []string{"-a"}

	logrus.Debug(
		fmt.Sprintf(
			"Getting LSB data using the following command : %s %s",
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
		}).Debug("Could not run command to list network interfaces")
		return lsbData, executionOutput.ErrorDetails
	}

	splittedOutput := strings.SplitSeq(executionOutput.Stdout, "\n")
	for line := range splittedOutput {
		if line != "" && strings.Contains(line, ":") {
			splittedLine := strings.Split(line, ":")
			if strings.Contains(splittedLine[0], "Distributor ID") {
				lsbData.DistributorId = strings.TrimSpace(splittedLine[1])
			} else if strings.Contains(splittedLine[0], "Description") {
				lsbData.Description = strings.TrimSpace(splittedLine[1])
			} else if strings.Contains(splittedLine[0], "Release") {
				lsbData.Release = strings.TrimSpace(splittedLine[1])
			} else if strings.Contains(splittedLine[0], "Codename") {
				lsbData.Codename = strings.TrimSpace(splittedLine[1])
			}
		}
	}

	return lsbData, nil
}
