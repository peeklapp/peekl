package user

import (
	"fmt"
	"os/user"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Ability to create the resources

type UserData struct {
	Username   string   `yaml:"username" json:"username" mapstructure:"username"`
	Groups     []string `yaml:"groups" json:"groups" mapstructure:"groups"`
	ManageHome bool     `yaml:"manage_home" json:"manage_home" mapstructure:"manage_home"`
	Shell      string   `yaml:"shell" json:"shell" mapstructure:"shell"`
}

type UserResource struct {
	Title   string
	Type    string
	Present bool
	Data    UserData
}

func (u *UserResource) exist() bool {
	logrus.Debug(
		fmt.Sprintf(
			"Checking if user (%s) exist using builtin Go package `os/user`",
			u.Data.Username,
		),
	)

	_, err := user.Lookup(u.Data.Username)
	if err != nil {
		logrus.Debug(fmt.Sprintf("User (%s) does not exist", u.Data.Username))
		return false
	}

	logrus.Debug(fmt.Sprintf("User (%s) exist", u.Data.Username))
	return true
}

func (u *UserResource) getCurrentGroups() ([]string, error) {
	var groups []string

	command := "id"
	args := []string{"-nG", u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"Getting user (%s) group using the following command : %s %s",
			u.Data.Username,
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
		}).Debug("Could not execute command to get user shell")
		return groups, executionOutput.ErrorDetails
	}

	for grp := range strings.FieldsSeq(strings.Trim(executionOutput.Stdout, "\n")) {
		groups = append(groups, grp)
	}

	return groups, nil
}

func (u *UserResource) addToGroup(group string) error {
	command := "adduser"
	args := []string{u.Data.Username, group}

	logrus.Debug(
		fmt.Sprintf(
			"Adding user (%s) to group (%s) using the following command : %s %s",
			u.Data.Username,
			group,
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
		}).Debug("")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (u *UserResource) addUserToGroupsIfNeeded() (bool, error) {
	var didSomething bool

	groups, err := u.getCurrentGroups()
	if err != nil {
		return didSomething, err
	}

	for _, group := range u.Data.Groups {
		if !slices.Contains(groups, group) {
			logrus.Info(
				fmt.Sprintf(
					"User (%s) is not a member of group (%s) but should be",
					u.Data.Username,
					group,
				),
			)
			err := u.addToGroup(group)
			if err != nil {
				return didSomething, err
			}
			logrus.Info(
				fmt.Sprintf(
					"User (%s) added to group (%s)",
					u.Data.Username,
					group,
				),
			)
			didSomething = true
		}
	}

	return didSomething, nil
}

func (u *UserResource) getCurrentShell() (string, error) {
	command := "getent"
	args := []string{"passwd", u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"Getting user (%s) shell using the following command : %s %s",
			u.Data.Username,
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
		}).Debug("Could not execute command to get user shell")
		return "", executionOutput.ErrorDetails
	}

	shell := strings.Trim(strings.Split(executionOutput.Stdout, ":")[6], "\n")

	logrus.Debug(
		fmt.Sprintf("Found shell (%s) for user (%s)", shell, u.Data.Username),
	)
	return shell, nil
}

func (u *UserResource) setShell() error {
	command := "chsh"
	args := []string{"-s", u.Data.Shell, u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"Updating user (%s) shell using the following command : %s %s",
			u.Data.Username,
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
		}).Debug("Could not execute command to change user shell")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (u *UserResource) updateShellIfNeeded() (bool, error) {
	var didSomething bool

	userShell, err := u.getCurrentShell()
	if err != nil {
		return didSomething, err
	}

	if userShell != u.Data.Shell {
		logrus.Info(
			fmt.Sprintf(
				"Shell for user (%s) should be (%s) but is (%s)",
				u.Data.Username,
				u.Data.Shell,
				userShell,
			),
		)
		err := u.setShell()
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"Shell for user (%s) has been updated from (%s) to (%s)",
				u.Data.Username,
				userShell,
				u.Data.Shell,
			),
		)
		didSomething = true
	}

	return didSomething, nil
}

func (u *UserResource) create() error {
	command := "adduser"
	args := []string{u.Data.Username}

	// Whether or not to create home directory
	if !u.Data.ManageHome {
		args = append(args, "--no-create-home")
	}

	// Add shell parameter to command
	args = append(args, "--shell")
	args = append(args, u.Data.Shell)

	logrus.Debug(
		fmt.Sprintf(
			"Creating user (%s) using the following command : %s %s",
			u.Data.Username,
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
		}).Debug("Could not execute command to create user")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (u *UserResource) delete() error {
	command := "deluser"
	args := []string{u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"Deleting user (%s) using the following command : %s %s",
			u.Data.Username,
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
		}).Debug("Could not execute command to delete user")
		return executionOutput.ErrorDetails
	}
	return nil
}

func (u *UserResource) Process() (models.ResourceResult, error) {
	var result models.ResourceResult

	// Handle user creation of deletion if needed
	if !u.exist() && u.Present {
		logrus.Info(
			fmt.Sprintf("User (%s) does not exist but should", u.Data.Username),
		)
		err := u.create()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("User (%s) created", u.Data.Username),
		)
		result.Created = true
	} else if u.exist() && !u.Present {
		logrus.Info(
			fmt.Sprintf("User (%s) exist but should not", u.Data.Username),
		)
		err := u.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("User (%s) deleted", u.Data.Username),
		)
		result.Deleted = true
	}

	// Make sure user correspond to resource
	if u.exist() && u.Present {
		// Update shell
		shellHasChanged, err := u.updateShellIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		// Updating user group membership
		userHasBeenAddedToAGroup, err := u.addUserToGroupsIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		if shellHasChanged || userHasBeenAddedToAGroup {
			result.Updated = true
		}
	}

	return result, nil
}

func (u *UserResource) String() string {
	return fmt.Sprintf("%s/%s", u.Type, u.Title)
}

func NewUserResource(resource *models.Resource) (*UserResource, error) {
	var userResource UserResource

	defaults := map[string]any{
		"shell":       "/bin/bash",
		"manage_home": true,
	}

	// Define data struct
	var userData UserData

	// First we set default values
	err := mapstructure.Decode(defaults, &userData)
	if err != nil {
		return &userResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(resource.Data, &userData)
	if err != nil {
		return &userResource, err
	}

	userResource.Title = resource.Title
	userResource.Type = resource.Type
	userResource.Present = resource.Present
	userResource.Data = userData

	return &userResource, nil
}
