package rpm

import (
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetInstalledPackagesList() (string, error) {
	command := "rpm"
	args := []string{"--query", "--all", "--queryformat", "%{NAME};%{VERSION}.%{RELEASE}\n"}

	logrus.Debug(
		fmt.Sprintf(
			"Getting list of installed packages using the following command : %s %s",
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
		}).Debug("Could not run command to list installed packages using rpm")
		return "", executionOutput.ErrorDetails
	}

	return executionOutput.Stdout, nil
}
