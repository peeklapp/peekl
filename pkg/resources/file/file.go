package file

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type FileData struct {
	Path        string      `mapstructure:"path"`
	Owner       string      `mapstructure:"owner"`
	Group       string      `mapstructure:"group"`
	Mode        fs.FileMode `mapstructure:"mode"`
	Source      string      `mapstructure:"source"`
	Content     string      `mapstructure:"content"`
	roleContext *models.RoleContext
}

type FileResource struct {
	resources.CommonFieldResource
	Data FileData
}

func (f *FileResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(f.Data.Path)
	if err != nil {
		return didSomething, err
	}

	// Update file permission if needed
	if stat.Mode() != f.Data.Mode {
		logrus.Info(
			fmt.Sprintf(
				"Mode for file (%s) should be (%s) but is (%s)",
				f.Data.Path,
				f.Data.Mode,
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
				f.Data.Mode,
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

func (f *FileResource) changeContentIfNeeded(content string) (bool, error) {
	var didSomething bool

	// First we open the file
	file, err := os.Open(f.Data.Path)
	if err != nil {
		return didSomething, err
	}
	defer file.Close()

	// Then we create MD5 object of file
	localFileHasher := md5.New()

	file.Seek(0, 0)
	if _, err := io.Copy(localFileHasher, file); err != nil {
		return didSomething, err
	}

	// Then we create MD5 object of content
	contentHasher := md5.New()
	io.WriteString(contentHasher, content)

	localFileMD5Value := hex.EncodeToString(localFileHasher.Sum(nil))
	contentMD5Value := hex.EncodeToString(contentHasher.Sum(nil))

	if localFileMD5Value != contentMD5Value {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Checksum for file (%s) should be (%s) but is (%s)",
				f.String(),
				f.Data.Path,
				contentMD5Value,
				localFileMD5Value,
			),
		)
		err := os.WriteFile(f.Data.Path, []byte(content), f.Data.Mode)
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"[%s] File (%s) content has been updated",
				f.String(),
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

func (f *FileResource) create(content string) error {
	file, err := os.Create(f.Data.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write content to file
	file.Write([]byte(content))

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
	err := os.Remove(f.Data.Path)
	return err
}

func (f *FileResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	var fileContent string
	var err error
	if f.Data.Source != "" {
		fileContent, err = context.ApiClient.RetrieveFile(f.Data.Source, context.Environment, f.Data.roleContext.RoleName)
		if err != nil {
			result.Failed = true
			return result, err
		}
	} else {
		fileContent = f.Data.Content
	}

	if !f.exist() && f.Present {
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) does not exist, but should", f.String(), f.Data.Path),
		)
		err := f.create(fileContent)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) created", f.String(), f.Data.Path),
		)
		result.Created = true
	} else if f.exist() && !f.Present {
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) exist, but should not", f.String(), f.Data.Path),
		)
		err := f.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) deleted", f.String(), f.Data.Path),
		)
		result.Deleted = true
	}

	// Process any other stuff
	if f.exist() && f.Present {
		var err error

		// Check content of the file
		var contentHasChanged bool
		contentHasChanged, err = f.changeContentIfNeeded(fileContent)
		if err != nil {
			result.Failed = true
			return result, err
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
	return fmt.Sprintf("%s / '%s'", f.Type, f.Title)
}

func (f *FileResource) When() string {
	return f.WhenCondition
}

func (f *FileResource) Register() string {
	return f.RegisterVariable
}

func (f *FileResource) Validate() error {
	validationErrors := []models.ValidationError{}

	fieldsThatCannotBeEmpty := [][]string{
		{f.Data.Path, "path"},
		{f.Data.Owner, "owner"},
		{f.Data.Group, "group"},
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

	if f.Data.Source != "" && f.Data.Content != "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "content / source",
				ViolatedRule: "content field and source field are mutually exclusive, you cannot use both.",
			},
		)
	}

	if f.Data.Source != "" && f.Data.roleContext == nil {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "source",
				ViolatedRule: "source field cannot be used outside of roles",
			},
		)
	}

	if len(validationErrors) > 0 {
		return models.ResourceValidationError{
			Type:             f.Type,
			Title:            f.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewFileResource(resource *models.Resource, dataField any, roleContext *models.RoleContext) (*FileResource, error) {
	var fileResource FileResource

	// Define defaults value
	defaults := map[string]any{
		"owner": "root",
		"group": "root",
		"mode":  0755,
	}

	// Define data struct
	var fileData FileData

	// First we set defaults values
	err := mapstructure.Decode(defaults, &fileData)
	if err != nil {
		return &fileResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(dataField, &fileData)
	if err != nil {
		return &fileResource, err
	}

	fileResource.Title = resource.Title
	fileResource.Type = resource.Type
	fileResource.Present = *resource.Present
	fileResource.WhenCondition = resource.When
	fileResource.RegisterVariable = resource.Register
	fileResource.Data = fileData
	fileResource.Data.roleContext = roleContext

	return &fileResource, nil
}
