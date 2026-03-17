package command

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/sirupsen/logrus"
)

type CommandData struct {
	Command        string   `mapstructure:"command"`
	Args           []string `mapstructure:"args"`
	Creates        string   `mapstructure:"creates"`
	RegisterOutput string   `mapstructure:"register_output"`
	Shell          string   `mapstructure:"shell"`
}

type CommandResource struct {
	resources.CommonFieldResource
	Data CommandData
}

func (c *CommandResource) createsAlreadyExist() bool {
	if _, err := os.Stat(c.Data.Creates); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (c *CommandResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	logrus.Debug(fmt.Sprintf("[%s] Checking if the command needs to be ran, depending on if the `creates` path exist.", c.String()))
	if c.createsAlreadyExist() {
		logrus.Debug(fmt.Sprintf("[%s] The command does not need to be ran, as the `creates` path exist.", c.String()))
		return result, nil
	}

	commandWithArgs := fmt.Sprintf("%s %s", c.Data.Command, strings.Join(c.Data.Args, " "))
	cmd := exec.Command(c.Data.Shell, "-c", commandWithArgs)

	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff

	logrus.Debug(fmt.Sprintf("[%s] Running the command : %s", c.String(), commandWithArgs))
	err := cmd.Run()

	if err != nil {
		result.Failed = true
		if exitError, ok := err.(*exec.ExitError); ok {
			logrus.WithFields(
				logrus.Fields{
					"command":   fmt.Sprintf("%s %s", c.Data.Command, strings.Join(c.Data.Args, " ")),
					"stderr":    stderrBuff.String(),
					"exit_code": exitError.ExitCode(),
				},
			).Error(fmt.Sprintf("[%s] Error during command execution", c.String()))
			return result, nil
		} else {
			logrus.Debug(fmt.Sprintf("[%s] Command failed in an unexpected way.", c.String()))
			return result, err
		}
	}

	if c.Data.RegisterOutput != "" {
		context.Variables[c.Data.RegisterOutput] = stdoutBuff.String()
	}

	result.Created = true
	return result, nil
}

func (c *CommandResource) String() string {
	return fmt.Sprintf("%s / '%s'", c.Type, c.Title)
}

func (c *CommandResource) When() string {
	return c.WhenCondition
}

func (c *CommandResource) Register() string {
	return c.RegisterVariable
}

func (c *CommandResource) Validate() error {
	validationErrors := []models.ValidationError{}

	fieldsThatCannotBeEmpty := [][]string{
		{c.Data.Command, "command"},
		{c.Data.Shell, "shell"},
	}
	for _, fieldToCheck := range fieldsThatCannotBeEmpty {
		if fieldToCheck[0] == "" {
			validationErrors = append(
				validationErrors,
				models.ValidationError{
					FieldName:    fieldToCheck[1],
					ViolatedRule: "Field cannot be empty",
				},
			)
		}
	}

	if c.Data.Shell != "" {
		switch c.Data.Shell {
		case "bash":
			break
		default:
			validationErrors = append(
				validationErrors,
				models.ValidationError{
					FieldName:    "shell",
					ViolatedRule: fmt.Sprintf("'%s' is not a valid shell", c.Data.Shell),
				},
			)
		}
	}

	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             c.Type,
			Title:            c.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewCommandResource(resource *models.Resource, dataField map[string]any, roleContext *models.RoleContext) (*CommandResource, error) {
	var commandResource CommandResource

	// Define defaults
	defaults := map[string]any{
		"shell": "bash",
	}

	// Define data struct
	var commandData CommandData

	// Set default values
	err := mapstructure.Decode(defaults, &commandData)
	if err != nil {
		return &commandResource, err
	}

	err = mapstructure.Decode(dataField, &commandData)
	if err != nil {
		return &commandResource, err
	}

	commandResource.Title = resource.Title
	commandResource.Type = resource.Type
	commandResource.Present = *resource.Present
	commandResource.WhenCondition = resource.When
	commandResource.RegisterVariable = resource.Register
	commandResource.Data = commandData

	return &commandResource, nil
}
