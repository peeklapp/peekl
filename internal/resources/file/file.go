package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/resources"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

type FileParameters struct {
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
	Parameters FileParameters
}

func (f *FileResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(f.Parameters.Path)
	if err != nil {
		return didSomething, err
	}

	// Update file permission if needed
	if stat.Mode() != f.Parameters.Mode {
		logrus.Info(
			fmt.Sprintf(
				"Mode for file (%s) should be (%s) but is (%s)",
				f.Parameters.Path,
				f.Parameters.Mode,
				stat.Mode(),
			),
		)
		if err := os.Chmod(f.Parameters.Path, f.Parameters.Mode); err != nil {
			return didSomething, err
		}
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Mode for file (%s) has been updated from (%s) to (%s)",
				f.Parameters.Path,
				stat.Mode(),
				f.Parameters.Mode,
			),
		)
	}

	return didSomething, nil
}

func (f *FileResource) changeOwnershipIfNeeded() (bool, error) {
	var didSomething bool

	stat, err := os.Stat(f.Parameters.Path)
	if err != nil {
		return didSomething, err
	}

	var foundUid int
	var foundGid int

	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		foundUid = int(stat.Uid)
		foundGid = int(stat.Gid)
	}

	expectedUid, err := utils.GetUserUidFromUsername(f.Parameters.Owner)
	if err != nil {
		return didSomething, err
	}
	expectedGid, err := utils.GetGroupGidFromName(f.Parameters.Group)
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
				f.Parameters.Path,
				f.Parameters.Owner,
				f.Parameters.Group,
				username,
				groupName,
			),
		)
		if err := os.Chown(f.Parameters.Path, expectedUid, expectedGid); err != nil {
			return didSomething, err
		}
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Ownership for file (%s) updated from (%s:%s) to (%s:%s)",
				f.Parameters.Path,
				username,
				groupName,
				f.Parameters.Owner,
				f.Parameters.Group,
			),
		)
	}

	return didSomething, nil
}

func (f *FileResource) changeContentIfNeeded(content string) (bool, error) {
	var didSomething bool

	// First we open the file
	file, err := os.Open(f.Parameters.Path)
	if err != nil {
		return didSomething, err
	}
	defer utils.CloseWithoutError(file)

	// Then we create MD5 object of file
	localFileHasher := md5.New()

	if _, err := file.Seek(0, 0); err != nil {
		return didSomething, err
	}

	if _, err := io.Copy(localFileHasher, file); err != nil {
		return didSomething, err
	}

	// Then we create MD5 object of content
	contentHasher := md5.New()
	_, err = io.WriteString(contentHasher, content)
	if err != nil {
		return didSomething, err
	}

	localFileMD5Value := hex.EncodeToString(localFileHasher.Sum(nil))
	contentMD5Value := hex.EncodeToString(contentHasher.Sum(nil))

	if localFileMD5Value != contentMD5Value {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Checksum for file (%s) should be (%s) but is (%s)",
				f.String(),
				f.Parameters.Path,
				contentMD5Value,
				localFileMD5Value,
			),
		)
		err := os.WriteFile(f.Parameters.Path, []byte(content), f.Parameters.Mode)
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"[%s] File (%s) content has been updated",
				f.String(),
				f.Parameters.Path,
			),
		)
		didSomething = true
	}

	return didSomething, nil
}

func (f *FileResource) exist() bool {
	return utils.FileExist(f.Parameters.Path, nil)
}

func (f *FileResource) create(content string) error {
	file, err := os.Create(f.Parameters.Path)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(file)

	// Write content to file
	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}

	// Change mode of file
	err = file.Chmod(f.Parameters.Mode)
	if err != nil {
		return err
	}

	// Set owner and group
	uid, err := utils.GetUserUidFromUsername(f.Parameters.Owner)
	if err != nil {
		return err
	}
	gid, err := utils.GetGroupGidFromName(f.Parameters.Group)
	if err != nil {
		return err
	}

	err = file.Chown(uid, gid)
	if err != nil {
		return err
	}

	return err
}

func (f *FileResource) delete() error {
	err := os.Remove(f.Parameters.Path)
	return err
}

func (f *FileResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	var fileContent string

	if f.Parameters.Source != "" {
		fullFilePath := filepath.Join(context.CodePath, "roles", f.Parameters.roleContext.RoleName, "files", f.Parameters.Source)
		if !utils.FileExist(fullFilePath, nil) {
			result.Failed = true
			return result, fmt.Errorf(
				"[%s] The file doesn't exist in role : %s",
				f.String(),
				filepath.Join(context.CodePath, "roles", f.Parameters.roleContext.RoleName, "files", f.Parameters.Source),
			)
		}
		rawFileContent, err := os.ReadFile(fullFilePath)
		if err != nil {
			result.Failed = true
			return result, err
		}
		fileContent = string(rawFileContent)
	} else {
		fileContent = f.Parameters.Content
	}

	if !f.exist() && f.Present {
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) does not exist, but should", f.String(), f.Parameters.Path),
		)
		err := f.create(fileContent)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) created", f.String(), f.Parameters.Path),
		)
		result.Created = true
	} else if f.exist() && !f.Present {
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) exist, but should not", f.String(), f.Parameters.Path),
		)
		err := f.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] File (%s) deleted", f.String(), f.Parameters.Path),
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
		{f.Parameters.Path, "path"},
		{f.Parameters.Owner, "owner"},
		{f.Parameters.Group, "group"},
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

	if f.Parameters.Source != "" && f.Parameters.Content != "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "content / source",
				ViolatedRule: "content field and source field are mutually exclusive, you cannot use both.",
			},
		)
	}

	if f.Parameters.Source != "" && f.Parameters.roleContext == nil {
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

func NewFileResource(resource *models.Resource, parametersField any, roleContext *models.RoleContext) (*FileResource, error) {
	var fileResource FileResource

	// Define defaults value
	defaults := map[string]any{
		"owner": "root",
		"group": "root",
		"mode":  0755,
	}

	// Define parameters struct
	var fileParameters FileParameters

	// First we set defaults values
	err := mapstructure.Decode(defaults, &fileParameters)
	if err != nil {
		return &fileResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(parametersField, &fileParameters)
	if err != nil {
		return &fileResource, err
	}

	fileResource.Title = resource.Title
	fileResource.Type = resource.Type
	fileResource.Present = *resource.Present
	fileResource.WhenCondition = resource.When
	fileResource.RegisterVariable = resource.Register
	fileResource.Parameters = fileParameters
	fileResource.Parameters.roleContext = roleContext

	return &fileResource, nil
}
