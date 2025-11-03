package group

import (
	"fmt"
	"os/user"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type GroupData struct {
	Name string `mapstructure:"name"`
}

type GroupResource struct {
	Title   string
	Type    string
	Present bool
	Data    GroupData
}

func (g *GroupResource) exist() bool {
	logrus.Debug(
		fmt.Sprintf(
			"Checking if group (%s) exist using the builtin Go package `os/user`",
			g.Data.Name,
		),
	)

	_, err := user.LookupGroup(g.Data.Name)
	if err != nil {
		logrus.Debug(fmt.Sprintf("Group (%s) does not exist", g.Data.Name))
		return false
	}

	logrus.Debug(fmt.Sprintf("Group (%s) exist", g.Data.Name))
	return true
}

func (g *GroupResource) create() error {
	command := "addgroup"
	args := []string{g.Data.Name}

	logrus.Debug(
		fmt.Sprintf(
			"Creating group (%s) using the following command : %s %s",
			g.Data.Name,
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
		}).Debug("Could not execute command to create group")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (g *GroupResource) delete() error {
	command := "delgroup"
	args := []string{g.Data.Name}

	logrus.Debug(
		fmt.Sprintf(
			"Deleting group (%s) using the following command : %s %s",
			g.Data.Name,
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
		}).Debug("Could not execute command to delete group")
		return executionOutput.ErrorDetails
	}

	return nil
}

func (g *GroupResource) Process(context *models.Context) (models.ResourceResult, error) {
	var result models.ResourceResult

	exist := g.exist()

	if !exist && g.Present {
		logrus.Info(
			fmt.Sprintf("Group (%s) does not exist but should", g.Data.Name),
		)
		err := g.create()
		if err != nil {
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Group (%s) created", g.Data.Name),
		)
		result.Created = true
	} else if exist && !g.Present {
		logrus.Info(
			fmt.Sprintf("Group (%s) exist but should not", g.Data.Name),
		)
		err := g.delete()
		if err != nil {
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Group (%s) deleted", g.Data.Name),
		)
		result.Deleted = true
	}

	return result, nil
}

func (g *GroupResource) String() string {
	return fmt.Sprintf("%s/%s", g.Type, g.Title)
}

func NewGroupResource(resource *models.Resource) (*GroupResource, error) {
	var groupResource GroupResource

	groupResource.Title = resource.Title
	groupResource.Type = resource.Type
	groupResource.Present = resource.Present

	err := mapstructure.Decode(resource.Data, &groupResource.Data)
	if err != nil {
		return &groupResource, err
	}

	return &groupResource, nil
}
