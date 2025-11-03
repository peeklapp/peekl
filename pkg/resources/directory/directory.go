package directory

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
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
	Title   string
	Type    string
	Present bool
	Data    DirectoryData
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

func (d *DirectoryResource) Process(context *models.Context) (models.ResourceResult, error) {
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
	return fmt.Sprintf("%s/%s", d.Type, d.Title)
}

func NewDirectoryResource(resource *models.Resource) (*DirectoryResource, error) {
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
	err = mapstructure.Decode(resource.Data, &directoryData)
	if err != nil {
		return &directoryResource, err
	}

	directoryResource.Title = resource.Title
	directoryResource.Type = resource.Type
	directoryResource.Present = resource.Present
	directoryResource.Data = directoryData

	return &directoryResource, nil
}
