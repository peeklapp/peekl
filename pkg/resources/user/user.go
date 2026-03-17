package user

import (
	"fmt"
	"os/user"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

// Ability to create the resources

type UserData struct {
	Username   string   `mapstructure:"username"`
	Groups     []string `mapstructure:"groups"`
	ManageHome bool     `mapstructure:"manage_home"`
	Shell      string   `mapstructure:"shell"`
}

type UserResource struct {
	resources.CommonFieldResource
	Data UserData
}

func (u *UserResource) exist() bool {
	logrus.Debug(
		fmt.Sprintf(
			"[%s] Checking if user (%s) exist using builtin Go package `os/user`",
			u.String(),
			u.Data.Username,
		),
	)

	_, err := user.Lookup(u.Data.Username)
	if err != nil {
		logrus.Debug(fmt.Sprintf("[%s] User (%s) does not exist", u.String(), u.Data.Username))
		return false
	}

	logrus.Debug(fmt.Sprintf("[%s] User (%s) exist", u.String(), u.Data.Username))
	return true
}

func (u *UserResource) getCurrentGroups() ([]string, error) {
	var groups []string

	command := "id"
	args := []string{"-nG", u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Getting user (%s) group using the following command : %s %s",
			u.String(),
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to get user shell", u.String()))
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
			"[%s] Adding user (%s) to group (%s) using the following command : %s %s",
			u.String(),
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
		}).Debug(fmt.Sprintf("[%s] Error while trying to add user to group", u.String()))
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
					"[%s] User (%s) is not a member of group (%s) but should be",
					u.String(),
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
					"[%s] User (%s) added to group (%s)",
					u.String(),
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
			"[%s] Getting user (%s) shell using the following command : %s %s",
			u.String(),
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to get user shell", u.String()))
		return "", executionOutput.ErrorDetails
	}

	shell := strings.Trim(strings.Split(executionOutput.Stdout, ":")[6], "\n")

	logrus.Debug(
		fmt.Sprintf("[%s] Found shell (%s) for user (%s)", shell, u.String(), u.Data.Username),
	)
	return shell, nil
}

func (u *UserResource) setShell() error {
	command := "chsh"
	args := []string{"-s", u.Data.Shell, u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Updating user (%s) shell using the following command : %s %s",
			u.String(),
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to change user shell", u.String()))
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
				"[%s] Shell for user (%s) should be (%s) but is (%s)",
				u.String(),
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
				"[%s] Shell for user (%s) has been updated from (%s) to (%s)",
				u.String(),
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
	command := "useradd"
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
			"[%s] Creating user (%s) using the following command : %s %s",
			u.String(),
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
		}).Info(fmt.Sprintf("[%s] Could not execute command to create user", u.String()))
		return executionOutput.ErrorDetails
	}
	return nil
}

func (u *UserResource) delete() error {
	command := "deluser"
	args := []string{u.Data.Username}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Deleting user (%s) using the following command : %s %s",
			u.String(),
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to delete user", u.String()))
		return executionOutput.ErrorDetails
	}
	return nil
}

func (u *UserResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	// Handle user creation of deletion if needed
	if !u.exist() && u.Present {
		logrus.Info(
			fmt.Sprintf("[%s] User (%s) does not exist but should", u.String(), u.Data.Username),
		)
		err := u.create()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] User (%s) created", u.String(), u.Data.Username),
		)
		result.Created = true
	} else if u.exist() && !u.Present {
		logrus.Info(
			fmt.Sprintf("[%s] User (%s) exist but should not", u.String(), u.Data.Username),
		)
		err := u.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] User (%s) deleted", u.String(), u.Data.Username),
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
	return fmt.Sprintf("%s / '%s'", u.Type, u.Title)
}

func (u *UserResource) When() string {
	return u.WhenCondition
}

func (u *UserResource) Register() string {
	return u.RegisterVariable
}

func (u *UserResource) Validate() error {
	validationErrors := []models.ValidationError{}

	// Check if the provided username is not empty
	if u.Data.Username == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "username",
				ViolatedRule: "Field cannot be empty",
			},
		)
	}

	// If any validation error, return error
	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             u.Type,
			Title:            u.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewUserResource(resource *models.Resource, dataField any, roleContext *models.RoleContext) (*UserResource, error) {
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
	err = mapstructure.Decode(dataField, &userData)
	if err != nil {
		return &userResource, err
	}

	userResource.Title = resource.Title
	userResource.Type = resource.Type
	userResource.Present = *resource.Present
	userResource.WhenCondition = resource.When
	userResource.RegisterVariable = resource.Register
	userResource.Data = userData

	return &userResource, nil
}
