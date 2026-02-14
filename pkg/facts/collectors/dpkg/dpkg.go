package dpkg

import (
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetInstalledPackagesList() ([]models.Package, error) {
	var pkgs []models.Package

	command := "dpkg-query"
	args := []string{"-W", "-f", "${Package};${Version}\n"}

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
		}).Debug("Could not run command to list installed packages using dpkg")
		return pkgs, executionOutput.ErrorDetails
	}

	splittedOutput := strings.SplitSeq(executionOutput.Stdout, "\n")
	for line := range splittedOutput {
		if line != "" {
			var pkg models.Package
			splittedLine := strings.Split(line, ";")
			pkg.Name = splittedLine[0]
			pkg.Version = splittedLine[1]
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}
