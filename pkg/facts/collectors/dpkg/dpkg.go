package dpkg

import (
	"fmt"
	"strings"

	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

func GetInstalledPackagesList() ([]models.PackageFact, error) {
	var pkgs []models.PackageFact

	command := "dpkg-query"
	args := []string{"-W", "-f", "${Package};${Version}\n"}

	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		})
	}

	splittedOutput := strings.SplitSeq(executionOutput.Stdout, "\n")
	for line := range splittedOutput {
		if line != "" {
			var pkg models.PackageFact
			splittedLine := strings.Split(line, ";")
			pkg.Name = splittedLine[0]
			pkg.Version = splittedLine[1]
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}
