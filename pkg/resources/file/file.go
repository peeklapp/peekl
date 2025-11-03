package file

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type FileData struct {
	Path        string      `yaml:"path" json:"path" mapstructure:"path"`
	Owner       string      `yaml:"owner" json:"owner" mapstructure:"owner"`
	Group       string      `yaml:"group" json:"group" mapstructure:"group"`
	Content     string      `yaml:"content" json:"content" mapstructure:"content"`
	Mode        fs.FileMode `yaml:"mode" json:"mode" mapstructure:"mode"`
	Directory   bool        `yaml:"directory" json:"directory" mapstructure:"directory"`
	ForceDelete bool        `yaml:"force_delete" json:"force_delete" mapstructure:"force_delete"`
}

type FileResource struct {
	Title   string
	Type    string
	Present bool
	Data    FileData
}

func (f *FileResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(f.Data.Path)
	if err != nil {
		return didSomething, err
	}

	updatedMode := f.Data.Mode
	if f.Data.Directory {
		updatedMode = f.Data.Mode | os.ModeDir
	}

	// Update file permission if needed
	if stat.Mode() != updatedMode {
		logrus.Info(
			fmt.Sprintf(
				"Mode for file (%s) should be (%s) but is (%s)",
				f.Data.Path,
				updatedMode,
				stat.Mode(),
			),
		)
		os.Chmod(f.Data.Path, f.Data.Mode)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Mode for file (%s) has been updated from (%s) to (%s)",
				f.Data.Path,
				stat.Mode(),
				updatedMode,
			),
		)
	}

	return didSomething, nil
}

func (f *FileResource) changeOwnershipIfNeeded() (bool, error) {
	var didSomething bool

	stat, err := os.Stat(f.Data.Path)
	if err != nil {
		return didSomething, err
	}

	var foundUid int
	var foundGid int

	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		foundUid = int(stat.Uid)
		foundGid = int(stat.Gid)
	}

	expectedUid, err := utils.GetUserUidFromUsername(f.Data.Owner)
	if err != nil {
		return didSomething, err
	}
	expectedGid, err := utils.GetGroupGidFromName(f.Data.Group)
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
				"Ownership for file (%s) should (%s:%s) but is (%s:%s)",
				f.Data.Path,
				f.Data.Owner,
				f.Data.Group,
				username,
				groupName,
			),
		)
		os.Chown(f.Data.Path, expectedUid, expectedGid)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Ownership for file (%s) updated from (%s:%s) to (%s:%s)",
				f.Data.Path,
				username,
				groupName,
				f.Data.Owner,
				f.Data.Group,
			),
		)
	}

	return didSomething, nil
}

func (f *FileResource) changeContentIfNeeded() (bool, error) {
	var didSomething bool

	// First we open the file
	file, err := os.Open(f.Data.Path)
	if err != nil {
		return didSomething, err
	}
	defer file.Close()

	// Then we create MD5 object of file
	fileMD5 := md5.New()
	if _, err := io.Copy(fileMD5, file); err != nil {
		return didSomething, err
	}

	// Then we create MD5 object of content
	contentMD5 := md5.New()
	io.WriteString(contentMD5, f.Data.Content)

	fileMD5Checksum := fmt.Sprintf("%x", fileMD5.Sum(nil))
	contentMD5Checksum := fmt.Sprintf("%x", contentMD5.Sum(nil))

	if fileMD5Checksum != contentMD5Checksum {
		logrus.Info(
			fmt.Sprintf(
				"Checksum for file (%s) should be (%s) but is (%s)",
				f.Data.Path,
				contentMD5Checksum,
				fileMD5Checksum,
			),
		)
		err := os.WriteFile(f.Data.Path, []byte(f.Data.Content), f.Data.Mode)
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"File (%s) content has been updated",
				f.Data.Path,
			),
		)
		didSomething = true
	}

	return didSomething, nil
}

func (f *FileResource) exist() bool {
	if _, err := os.Stat(f.Data.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (f *FileResource) create() error {
	if f.Data.Directory {
		err := os.Mkdir(f.Data.Path, f.Data.Mode)
		if err != nil {
			return err
		}
		return nil
	}

	file, err := os.Create(f.Data.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write content to file
	file.Write([]byte(f.Data.Content))

	// Change mode of file
	file.Chmod(f.Data.Mode)

	// Set owner and group
	uid, err := utils.GetUserUidFromUsername(f.Data.Owner)
	if err != nil {
		return err
	}
	gid, err := utils.GetGroupGidFromName(f.Data.Group)
	if err != nil {
		return err
	}
	file.Chown(uid, gid)

	return nil
}

func (f *FileResource) delete() error {
	if f.Data.ForceDelete {
		err := os.RemoveAll(f.Data.Path)
		return err
	}
	err := os.Remove(f.Data.Path)
	return err
}

func (f *FileResource) Process() (models.ResourceResult, error) {
	var result models.ResourceResult

	if !f.exist() && f.Present {
		logrus.Info(
			fmt.Sprintf("File (%s) does not exist, but should", f.Data.Path),
		)
		err := f.create()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("File (%s) created", f.Data.Path),
		)
		result.Created = true
	} else if f.exist() && !f.Present {
		logrus.Info(
			fmt.Sprintf("File (%s) exist, but should not", f.Data.Path),
		)
		err := f.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("File (%s) deleted", f.Data.Path),
		)
		result.Deleted = true
	}

	// Process any other stuff
	if f.exist() && f.Present {
		var err error

		// Check content of the file
		var contentHasChanged bool
		if !f.Data.Directory {
			contentHasChanged, err = f.changeContentIfNeeded()
			if err != nil {
				result.Failed = true
				return result, err
			}
		}

		// Check permissions of the file
		permissionsHasBeenChanged, err := f.changePermissionsIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		// Check owner/group of the file
		ownershipHasBeenChanged, err := f.changeOwnershipIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		if permissionsHasBeenChanged || ownershipHasBeenChanged || contentHasChanged {
			result.Updated = true
		}
	}

	return result, nil
}

func (f *FileResource) String() string {
	return fmt.Sprintf("%s/%s", f.Type, f.Title)
}

func NewFileResource(resource *models.Resource) (*FileResource, error) {
	var fileResource FileResource

	// Define defaults value
	defaults := map[string]any{
		"owner":        "root",
		"group":        "root",
		"mode":         0755,
		"force_delete": false,
	}

	// Define data struct
	var fileData FileData

	// First we set defaults values
	err := mapstructure.Decode(defaults, &fileData)
	if err != nil {
		return &fileResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(resource.Data, &fileData)
	if err != nil {
		return &fileResource, err
	}

	fileResource.Title = resource.Title
	fileResource.Type = resource.Type
	fileResource.Present = resource.Present
	fileResource.Data = fileData

	return &fileResource, nil
}
