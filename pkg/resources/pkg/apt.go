package pkg

import (
	"fmt"
	"strings"

	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type AptInstaller struct{}

func (a AptInstaller) Install(packageName string, version string) error {
	command := "apt-get"
	args := []string{"install", "-y"}

	if version != "" {
		args = append(args, fmt.Sprintf("%s=%s", packageName, version))
	} else {
		args = append(args, packageName)
	}

	logrus.Debug(
		fmt.Sprintf(
			"Installing package (%s) using the following command : %s %s",
			packageName,
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
		}).Debug("Could not execute command to install package")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (a AptInstaller) Remove(packageName string) error {
	command := "apt-get"
	args := []string{"remove", "-y", packageName}

	logrus.Debug(
		fmt.Sprintf(
			"Removing package (%s) using the following command : %s %s",
			packageName,
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
		}).Debug("Could not execute command to remove package")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (a AptInstaller) IsPackageInstalled(packageName string) bool {
	command := "dpkg"
	args := []string{"-s", packageName}

	logrus.Debug(
		fmt.Sprintf(
			"Checking if package (%s) is installed using the following command : %s %s",
			packageName,
			command,
			strings.Join(args, " "),
		),
	)

	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		return false
	}
	return true
}
