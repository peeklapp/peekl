package group

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/resources"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

type GroupParameters struct {
	Name string `mapstructure:"name"`
}

type GroupResource struct {
	resources.CommonFieldResource
	Parameters GroupParameters
}

func (g *GroupResource) exist() bool {
	logrus.Debug(
		fmt.Sprintf(
			"[%s] Checking if group (%s) exist using the builtin Go package `os/user`",
			g.String(),
			g.Parameters.Name,
		),
	)

	_, err := user.LookupGroup(g.Parameters.Name)
	if err != nil {
		logrus.Debug(fmt.Sprintf("[%s] Group (%s) does not exist", g.String(), g.Parameters.Name))
		return false
	}

	logrus.Debug(fmt.Sprintf("[%s] Group (%s) exist", g.String(), g.Parameters.Name))
	return true
}

func (g *GroupResource) create() error {
	command := "addgroup"
	args := []string{g.Parameters.Name}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Creating group (%s) using the following command : %s %s",
			g.String(),
			g.Parameters.Name,
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to create group", g.String()))
		return executionOutput.ErrorDetails
	}

	return nil
}

func (g *GroupResource) delete() error {
	command := "delgroup"
	args := []string{g.Parameters.Name}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Deleting group (%s) using the following command : %s %s",
			g.String(),
			g.Parameters.Name,
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to delete group", g.String()))
		return executionOutput.ErrorDetails
	}

	return nil
}

func (g *GroupResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	exist := g.exist()

	if !exist && g.Present {
		logrus.Info(
			fmt.Sprintf("[%s] Group (%s) does not exist but should", g.String(), g.Parameters.Name),
		)
		err := g.create()
		if err != nil {
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Group (%s) created", g.String(), g.Parameters.Name),
		)
		result.Created = true
	} else if exist && !g.Present {
		logrus.Info(
			fmt.Sprintf("[%s] Group (%s) exist but should not", g.String(), g.Parameters.Name),
		)
		err := g.delete()
		if err != nil {
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Group (%s) deleted", g.String(), g.Parameters.Name),
		)
		result.Deleted = true
	}

	return result, nil
}

func (g *GroupResource) String() string {
	return fmt.Sprintf("%s / '%s'", g.Type, g.Title)
}

func (g *GroupResource) When() string {
	return g.WhenCondition
}

func (g *GroupResource) Register() string {
	return g.RegisterVariable
}

func (g *GroupResource) Validate() error {
	validationErrors := []models.ValidationError{}

	// Check if the provided name is not empty
	if g.Parameters.Name == "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "name",
				ViolatedRule: "Field cannot be empty",
			},
		)
	}

	// If any validation error, return error
	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             g.Type,
			Title:            g.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewGroupResource(resource *models.Resource, parametersField any, roleContext *models.RoleContext) (*GroupResource, error) {
	var groupResource GroupResource

	groupResource.Title = resource.Title
	groupResource.Type = resource.Type
	groupResource.Present = *resource.Present
	groupResource.WhenCondition = resource.When
	groupResource.RegisterVariable = resource.Register

	err := mapstructure.Decode(parametersField, &groupResource.Parameters)
	if err != nil {
		return &groupResource, err
	}

	return &groupResource, nil
}
