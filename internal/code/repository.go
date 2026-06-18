package code

import (
	"fmt"
	"os"
	"strings"

	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

func startSshAgent() error {
	command := "ssh-agent"
	args := []string{"-s"}

	logrus.Debug(fmt.Sprintf("Starting an SSH agent with the following command : %s %s", command, strings.Join(args, " ")))
	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug("Could not execute command to start ssh agent")
		return executionOutput.ErrorDetails
	}

	for _, line := range strings.Split(executionOutput.Stdout, "\n") {
		if strings.Contains(line, "SSH_AUTH_SOCK=") {
			splittedLineOnSemicolon := strings.Split(line, ";")
			splittedLineOnEqualSign := strings.Split(splittedLineOnSemicolon[0], "=")
			err := os.Setenv("SSH_AUTH_SOCK", splittedLineOnEqualSign[1])
			if err != nil {
				return err
			}
		} else if strings.Contains(line, "SSH_AGENT_PID=") {
			splittedLineOnSemicolon := strings.Split(line, ";")
			splittedLineOnEqualSign := strings.Split(splittedLineOnSemicolon[0], "=")
			err := os.Setenv("SSH_AGENT_PID", splittedLineOnEqualSign[1])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func addKeyToSshAgent(keyPath string) error {
	command := "ssh-add"
	args := []string{keyPath}

	logrus.Debug(fmt.Sprintf("Adding an SSH key to SSH agent with the following command : %s %s", command, strings.Join(args, " ")))
	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug("Could not execute command to start ssh agent")
		return executionOutput.ErrorDetails
	}

	return nil
}

func killSshAgent() error {
	command := "ssh-agent"
	args := []string{"-k"}

	logrus.Debug(fmt.Sprintf("Killing SSH agent with the following command : %s %s", command, strings.Join(args, " ")))
	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug("Could not execute command to start ssh agent")
		return executionOutput.ErrorDetails
	}

	return nil

}

func cloneRepository(conf *config.RepositoryConfig, branch string, folderPath string) error {
	doSsh := !strings.HasPrefix(conf.Url, "http")
	if doSsh {
		if err := startSshAgent(); err != nil {
			return err
		}
		if err := addKeyToSshAgent(conf.Key); err != nil {
			return err
		}
	}

	command := "git"
	args := []string{"clone", conf.Url, "--branch", branch, folderPath, "--depth", "1"}

	logrus.Info(fmt.Sprintf("Cloning repository on branch '%s'", branch))
	logrus.Debug(fmt.Sprintf("Cloning repository with following command : %s %s", command, strings.Join(args, " ")))
	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode != 0 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug("Could not execute command to clone directory")
		return executionOutput.ErrorDetails
	}
	logrus.Info("Finshed cloning repository")

	if doSsh {
		if err := killSshAgent(); err != nil {
			return err
		}
	}

	return nil
}
