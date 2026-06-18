package collectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

func getRawInterfacesWithIp() (string, error) {
	command := "ip"
	args := []string{"-j", "link"}

	logrus.Debug(
		fmt.Sprintf(
			"Getting list of network interfaces using the following command : %s %s",
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
		return "", executionOutput.ErrorDetails
	}

	return executionOutput.Stdout, nil
}

func processIpOutput(rawInterfaces string) ([]models.NetworkInterface, error) {
	var networkInterfaces []models.NetworkInterface
	err := json.Unmarshal([]byte(rawInterfaces), &networkInterfaces)
	if err != nil {
		return networkInterfaces, fmt.Errorf("An error happened while deserializing network interfaces data : %s", err.Error())
	}
	return networkInterfaces, nil
}

func GetNetworkInterfaces() ([]models.NetworkInterface, error) {
	rawData, err := getRawInterfacesWithIp()
	if err != nil {
		return nil, err
	}
	return processIpOutput(rawData)
}
