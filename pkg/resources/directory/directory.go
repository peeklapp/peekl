package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type DirectoryData struct {
	Path        string      `mapstructure:"path"`
	Owner       string      `mapstructure:"owner"`
	Group       string      `mapstructure:"group"`
	Mode        fs.FileMode `mapstructure:"mode"`
	ForceDelete bool        `mapstructure:"force_delete"`
}

type DirectoryResource struct {
	resources.CommonFieldResource
	Data DirectoryData
}

func (d *DirectoryResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(d.Data.Path)
	if err != nil {
		return didSomething, err
	}

	updatedMode := d.Data.Mode | os.ModeDir

	// Update file permission if needed
	if stat.Mode() != updatedMode {
		logrus.Info(
			fmt.Sprintf(
				"Mode for directory (%s) should be (%s) but is (%s)",
				d.Data.Path,
				updatedMode,
				stat.Mode(),
			),
		)
		os.Chmod(d.Data.Path, d.Data.Mode)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Mode for directory (%s) has been updated from (%s) to (%s)",
				d.Data.Path,
				stat.Mode(),
				updatedMode,
			),
		)
	}

	return didSomething, nil
}

func (d *DirectoryResource) changeOwnershipIfNeeded() (bool, error) {
	var didSomething bool

	stat, err := os.Stat(d.Data.Path)
	if err != nil {
		return didSomething, err
	}

	var foundUid int
	var foundGid int

	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		foundUid = int(stat.Uid)
		foundGid = int(stat.Gid)
	}

	expectedUid, err := utils.GetUserUidFromUsername(d.Data.Owner)
	if err != nil {
		return didSomething, err
	}
	expectedGid, err := utils.GetGroupGidFromName(d.Data.Group)
	if err != nil {
		return didSomething, err
	}

	if expectedUid != foundUid || expectedGid != foundGid {
		username, err := utils.GetUserUsernameFromUid(foundUid)
		if err != nil {
			return didSomething, err
		}
		groupName, err := utils.GetGroupNameFromGid(foundGid)
		if err != nil {
			return didSomething, err
		}

		logrus.Info(
			fmt.Sprintf(
				"Ownership for directory (%s) should (%s:%s) but is (%s:%s)",
				d.Data.Path,
				d.Data.Owner,
				d.Data.Group,
				username,
				groupName,
			),
		)
		os.Chown(d.Data.Path, expectedUid, expectedGid)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Ownership for directory (%s) updated from (%s:%s) to (%s:%s)",
				d.Data.Path,
				username,
				groupName,
				d.Data.Owner,
				d.Data.Group,
			),
		)
	}

	return didSomething, nil
}

func (d *DirectoryResource) exist() bool {
	if _, err := os.Stat(d.Data.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (d *DirectoryResource) create() error {
	err := os.Mkdir(d.Data.Path, d.Data.Mode)
	if err != nil {
		return err
	}

	return nil
}

func (d *DirectoryResource) delete() error {
	if d.Data.ForceDelete {
		err := os.RemoveAll(d.Data.Path)
		return err
	}
	err := os.Remove(d.Data.Path)
	return err
}

func (d *DirectoryResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	if !d.exist() && d.Present {
		logrus.Info(
			fmt.Sprintf("Directory (%s) does not exist, but should", d.Data.Path),
		)
		err := d.create()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Directory (%s) created", d.Data.Path),
		)
		result.Created = true
	} else if d.exist() && !d.Present {
		logrus.Info(
			fmt.Sprintf("Directory (%s) exist, but should not", d.Data.Path),
		)
		err := d.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Directory (%s) deleted", d.Data.Path),
		)
		result.Deleted = true
	}

	// Process any other stuff
	if d.exist() && d.Present {
		var err error

		// Check permissions of the file
		permissionsHasBeenChanged, err := d.changePermissionsIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		// Check owner/group of the file
		ownershipHasBeenChanged, err := d.changeOwnershipIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		if permissionsHasBeenChanged || ownershipHasBeenChanged {
			result.Updated = true
		}
	}

	return result, nil
}

func (d *DirectoryResource) String() string {
	return fmt.Sprintf("%s / '%s'", d.Type, d.Title)
}

func (d *DirectoryResource) When() string {
	return d.WhenCondition
}

func (d *DirectoryResource) Register() string {
	return d.RegisterVariable
}

func (d *DirectoryResource) Validate() error {
	validationErrors := []models.ValidationError{}

	fieldsThatCannotBeEmpty := [][]string{
		{d.Data.Path, "path"},
		{d.Data.Owner, "owner"},
		{d.Data.Group, "group"},
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

	// If any validation error, return error
	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             d.Type,
			Title:            d.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewDirectoryResource(resource *models.Resource, dataField any) (*DirectoryResource, error) {
	var directoryResource DirectoryResource

	// Define defaults value
	defaults := map[string]any{
		"owner":        "root",
		"group":        "root",
		"mode":         0644,
		"force_delete": false,
	}

	// Define data struct
	var directoryData DirectoryData

	// First we set defaults values
	err := mapstructure.Decode(defaults, &directoryData)
	if err != nil {
		return &directoryResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(dataField, &directoryData)
	if err != nil {
		return &directoryResource, err
	}

	directoryResource.Title = resource.Title
	directoryResource.Type = resource.Type
	directoryResource.Present = *resource.Present
	directoryResource.WhenCondition = resource.When
	directoryResource.RegisterVariable = resource.Register
	directoryResource.Data = directoryData

	return &directoryResource, nil
}
