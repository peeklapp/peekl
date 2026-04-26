package systemdDaemon

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type SystemdDaemonData struct {
	Reload bool `mapstructure:"reload"`
}

type SystemdDaemonResource struct {
	resources.CommonFieldResource
	Data SystemdDaemonData
}

func (s *SystemdDaemonResource) reloadSystemdDaemon() error {
	command := "systemctl"
	args := []string{"daemon-reload"}

	logrus.Debug(
		fmt.Sprintf(
			"[%s] Reload systemd daemon using the following command : %s %s",
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
		}).Debug(fmt.Sprintf("[%s] Could not execute command to reload systemd daemon.", s.String()))
		return executionOutput.ErrorDetails
	}

	return nil
}

func (s *SystemdDaemonResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	if s.Data.Reload {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Systemd daemon should be reloaded",
				s.String(),
			),
		)
		err := s.reloadSystemdDaemon()
		if err != nil {
			result.Failed = true
			return result, err
		}
		result.Updated = true
		logrus.Info(
			fmt.Sprintf(
				"[%s] Systemd daemon has been reloaded",
				s.String(),
			),
		)
	}

	return result, nil
}

func (s *SystemdDaemonResource) String() string {
	return fmt.Sprintf("%s / '%s'", s.Type, s.Title)
}

func (s *SystemdDaemonResource) When() string {
	return s.WhenCondition
}

func (s *SystemdDaemonResource) Register() string {
	return s.RegisterVariable
}

func (s *SystemdDaemonResource) Validate() error {
	// Not that much to validate in this case
	return nil
}

func NewSystemdDaemonResource(resource *models.Resource, dataField any, roleContext *models.RoleContext) (*SystemdDaemonResource, error) {
	var systemdDaemonResource SystemdDaemonResource

	systemdDaemonResource.Title = resource.Title
	systemdDaemonResource.Type = resource.Type
	systemdDaemonResource.Present = *resource.Present
	systemdDaemonResource.WhenCondition = resource.When
	systemdDaemonResource.RegisterVariable = resource.Register

	err := mapstructure.Decode(dataField, &systemdDaemonResource.Data)
	if err != nil {
		return &systemdDaemonResource, err
	}

	return &systemdDaemonResource, nil
}
