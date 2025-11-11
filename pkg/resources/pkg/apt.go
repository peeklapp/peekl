package pkg

import (
	"fmt"
	"strings"

	"github.com/redat00/peekl/pkg/facts/collectors/dpkg"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type AptInstaller struct{}

func (a AptInstaller) Install(pkgs []models.Package) error {
	command := "apt-get"
	args := []string{"install", "-y"}

	for _, pkg := range pkgs {
		if pkg.Version != "" {
			args = append(args, fmt.Sprintf("%s=%s", pkg.Name, pkg.Version))
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
		}).Debug("Could not execute command to install package")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (a AptInstaller) Remove(pkgs []models.Package) error {
	command := "apt-get"
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
		}).Debug("Could not execute command to remove package")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (a AptInstaller) Upgrade(pkgs []models.Package) error {
	command := "apt-get"
	args := []string{"install", "-y", "--allow-downgrades"}

	for _, pkg := range pkgs {
		if pkg.Version != "" {
			args = append(args, fmt.Sprintf("%s=%s", pkg.Name, pkg.Version))
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
		}).Debug("Could not execute command to upgrade/downgrade package")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (a AptInstaller) ListInstalledPackages() ([]models.Package, error) {
	return dpkg.GetInstalledPackagesList()
}
