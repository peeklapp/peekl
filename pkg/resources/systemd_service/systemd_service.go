package systemdService

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

// This module and the way it works, while not being directly the same code,
// is inspired by the systemd_service module that you can find in Ansible.
//
// You can find their original code by following this link to Github :
// https://github.com/ansible/ansible/blob/devel/lib/ansible/modules/systemd_service.py

type SystemdServiceData struct {
	Name    string `mapstructure:"name"`
	Enabled bool   `mapstructure:"enabled"`
	Masked  bool   `mapstructure:"masked"`
	State   string `mapstructure:"state"`
}

type SystemdServiceResource struct {
	resources.CommonFieldResource
	Data SystemdServiceData
}

func (s *SystemdServiceResource) checkIfServiceIsEnabledOrMasked(checking string) (bool, error) {
	switch checking {
	case "enabled":
		break
	case "masked":
		break
	default:
		return false, fmt.Errorf("[%s] You can only checked for enabled of for masked", s.String())
	}

	command := "systemctl"
	args := []string{"is-enabled", s.Data.Name}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Checking if service is %s using the following command : %s %s",
			s.String(),
			checking,
			command,
			strings.Join(args, " "),
		),
	)

	executionOutput := utils.Execute(command, args...)
	if executionOutput.ErrorDetails.ExitCode > 1 {
		logrus.WithFields(logrus.Fields{
			"command":   fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
			"stderr":    executionOutput.ErrorDetails.Stderr,
			"exit_code": executionOutput.ErrorDetails.ExitCode,
		}).Debug(fmt.Sprintf("[%s] Could not execute command to verify if a service is enabled", s.String()))
		return false, executionOutput.ErrorDetails
	}

	if strings.Contains(executionOutput.Stdout, checking) {
		return true, nil
	}
	return false, nil
}

func (s *SystemdServiceResource) getServiceDetails() (map[string]string, error) {
	command := "systemctl"
	args := []string{"show", s.Data.Name}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Getting details of a systemd service using the following command : %s %s",
			s.String(),
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to get service details", s.String()))
		return nil, executionOutput.ErrorDetails
	}

	parsed := map[string]string{}
	splittedOutput := strings.SplitSeq(executionOutput.Stdout, "\n")
	for line := range splittedOutput {
		if line != "" {
			splittedLine := strings.Split(line, "=")
			parsed[splittedLine[0]] = splittedLine[1]
		}
	}

	return parsed, nil
}

func (s *SystemdServiceResource) doActionOnService(action string) error {
	switch action {
	case "enable":
		break
	case "disable":
		break
	case "mask":
		break
	case "unmask":
		break
	case "start":
		break
	case "restart":
		break
	case "stop":
		break
	case "reload":
		break
	default:
		return fmt.Errorf("unknown action %s", action)
	}

	command := "systemctl"
	args := []string{action, s.Data.Name}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Performing action (%s) on service (%s) with following command : %s %s",
			s.String(),
			action,
			s.Data.Name,
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to perform action on service", s.String()))
		return executionOutput.ErrorDetails
	}

	return nil
}

func (s *SystemdServiceResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	// Append .service at end of name if not already present
	if !strings.HasSuffix(s.Data.Name, ".service") {
		s.Data.Name = fmt.Sprintf("%s.service", s.Data.Name)
	}

	// Get raw details of our specific unit
	serviceDetails, err := s.getServiceDetails()
	if err != nil {
		result.Failed = true
		return result, err
	}

	// Check if unit is masked, unmask or mask if it needs to
	serviceMasked, err := s.checkIfServiceIsEnabledOrMasked("masked")
	if err != nil {
		result.Failed = true
		return result, err
	}

	if serviceMasked != s.Data.Masked {
		if s.Data.Masked {
			err := s.doActionOnService("mask")
			if err != nil {
				result.Failed = true
				return result, err
			}
		} else {
			err := s.doActionOnService("unmask")
			if err != nil {
				result.Failed = true
				return result, err
			}
		}
	}

	acceptableRunningStates := []string{"active", "activating"}
	switch s.Data.State {
	case "running":
		// Check if the service active state is considered running
		if !slices.Contains(acceptableRunningStates, serviceDetails["ActiveState"]) {
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) is not running but should", s.String(), s.Data.Name),
			)
			// Start the service
			err := s.doActionOnService("start")
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Updated = true
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) started", s.String(), s.Data.Name),
			)
		}
	case "stopped":
		if slices.Contains(acceptableRunningStates, serviceDetails["ActiveState"]) && serviceDetails["ActiveState"] != "deactivating" {
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) is running but should not", s.String(), s.Data.Name),
			)
			// Stop the service
			err := s.doActionOnService("stop")
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Updated = true
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) stopped", s.String(), s.Data.Name),
			)
		}
	case "restarted":
		// In case of a restart, we actually don't even have
		// to bother for any previous status of the service,
		// we simply send a restart action to the service
		logrus.Info(
			fmt.Sprintf("[%s] Service (%s) should be restarted", s.String(), s.Data.Name),
		)
		err := s.doActionOnService("restart")
		if err != nil {
			result.Failed = true
			return result, err
		}
		result.Updated = true
		logrus.Info(
			fmt.Sprintf("[%s] Service (%s) restarted", s.String(), s.Data.Name),
		)
	case "reloaded":
		logrus.Info(
			fmt.Sprintf("[%s] Service (%s) should be reloaded", s.String(), s.Data.Name),
		)
		err := s.doActionOnService("reload")
		if err != nil {
			result.Failed = true
			return result, err
		}
		result.Updated = true
		logrus.Info(
			fmt.Sprintf("[%s] Service (%s) reloaded", s.String(), s.Data.Name),
		)
	}

	// Check if service is enabled, enable if it needs to
	enabled, err := s.checkIfServiceIsEnabledOrMasked("enabled")
	if err != nil {
		result.Failed = true
		return result, err
	}
	if enabled != s.Data.Enabled {
		if s.Data.Enabled {
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) is not enabled but should", s.String(), s.Data.Name),
			)
			err := s.doActionOnService("enable")
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Updated = true
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) enabled", s.String(), s.Data.Name),
			)
		} else {
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) enabled but should not", s.String(), s.Data.Name),
			)
			err := s.doActionOnService("disable")
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Updated = true
			logrus.Info(
				fmt.Sprintf("[%s] Service (%s) disabled", s.String(), s.Data.Name),
			)
		}
	}

	return result, nil
}

func (s *SystemdServiceResource) String() string {
	return fmt.Sprintf("%s / '%s'", s.Type, s.Title)
}

func (s *SystemdServiceResource) When() string {
	return s.WhenCondition
}

func (s *SystemdServiceResource) Register() string {
	return s.RegisterVariable
}

func (s *SystemdServiceResource) Validate() error {
	validationErrors := []models.ValidationError{}

	fieldsThatCannotBeEmpty := [][]string{
		{s.Data.Name, "name"},
		{s.Data.State, "state"},
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

	if s.Data.State != "" {
		switch s.Data.State {
		case "running":
			break
		case "stopped":
			break
		case "restarted":
			break
		case "reloaded":
			break
		default:
			validationErrors = append(
				validationErrors,
				models.ValidationError{
					FieldName:    "state",
					ViolatedRule: fmt.Sprintf("'%s' is not a valid state", s.Data.State),
				},
			)
		}
	}

	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             s.Type,
			Title:            s.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewSystemdServiceResource(resource *models.Resource, dataField map[string]any, roleContext *models.RoleContext) (*SystemdServiceResource, error) {
	var systemdServiceResource SystemdServiceResource

	defaults := map[string]any{
		"enabled": true,
		"masked":  false,
	}

	var systemdServiceData SystemdServiceData

	// First we set default values
	err := mapstructure.Decode(defaults, &systemdServiceData)
	if err != nil {
		return &systemdServiceResource, err
	}

	// Then we override with the actual values
	err = mapstructure.Decode(dataField, &systemdServiceData)
	if err != nil {
		return &systemdServiceResource, err
	}

	systemdServiceResource.Title = resource.Title
	systemdServiceResource.Type = resource.Type
	systemdServiceResource.Present = *resource.Present
	systemdServiceResource.WhenCondition = resource.When
	systemdServiceResource.RegisterVariable = resource.Register
	systemdServiceResource.Data = systemdServiceData

	return &systemdServiceResource, nil
}
