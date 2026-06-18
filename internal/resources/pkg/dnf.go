package pkg

import (
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/internal/facts/collectors"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

type DnfInstaller struct{}

func (d DnfInstaller) Install(pkgs []models.Package) error {
	command := "dnf"
	args := []string{"install", "-y"}

	for _, pkg := range pkgs {
		if pkg.Version != "" {
			args = append(args, fmt.Sprintf("%s-%s", pkg.Name, pkg.Version))
		} else {
			args = append(args, pkg.Name)
		}
	}

	logrus.Debug(
		fmt.Sprintf(
			"Installing packages using the following command : %s %s",
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
		}).Debug("Could not execute command to install packages")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (d DnfInstaller) Remove(pkgs []models.Package) error {
	command := "dnf"
	args := []string{"remove", "-y"}

	for _, pkg := range pkgs {
		args = append(args, pkg.Name)
	}

	logrus.Debug(
		fmt.Sprintf(
			"Removing packages using the following command : %s %s",
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
		}).Debug("Could not execute command to remove packages")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (d DnfInstaller) Upgrade(pkgs []models.Package) error {
	command := "dnf"
	args := []string{"install", "-y"}

	for _, pkg := range pkgs {
		if pkg.Version != "" {
			args = append(args, fmt.Sprintf("%s-%s", pkg.Name, pkg.Version))
		} else {
			args = append(args, pkg.Name)
		}
	}

	logrus.Debug(
		fmt.Sprintf(
			"Upgrading/Downgrading packages using the following command : %s %s",
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
		}).Debug("Could not execute command to upgrade/downgrade packages")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (d DnfInstaller) ListInstalledPackages() ([]models.Package, error) {
	return collectors.GetPackagesByCollector("rpm")
}
