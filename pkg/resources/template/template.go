package template

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"maps"
	"os"
	"syscall"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/facts"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type TemplateData struct {
	Source      string         `mapstructure:"source"`
	Path        string         `mapstructure:"path"`
	Owner       string         `mapstructure:"owner"`
	Group       string         `mapstrucutre:"group"`
	Mode        fs.FileMode    `mapstructure:"mode"`
	Content     string         `mapstructure:"content"`
	Variables   map[string]any `mapstructure:"variables"`
	roleContext *models.RoleContext
}

type TemplateResource struct {
	resources.CommonFieldResource
	Data TemplateData
}

func (t *TemplateResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(t.Data.Path)
	if err != nil {
		return didSomething, err
	}

	// Update file permissions if needed
	if stat.Mode() != t.Data.Mode {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Mode for the template file (%s) should be (%s) but is (%s)",
				t.String(),
				t.Data.Path,
				t.Data.Mode,
				stat.Mode(),
			),
		)
		os.Chmod(t.Data.Path, t.Data.Mode)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"[%s] Mode for the template file (%s) has been updated from (%s) to (%s)",
				t.String(),
				t.Data.Path,
				stat.Mode(),
				t.Data.Mode,
			),
		)
	}
	return didSomething, nil
}

func (t *TemplateResource) changeOwnershipIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(t.Data.Path)
	if err != nil {
		return didSomething, err
	}

	var foundUid int
	var foundGid int

	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		foundUid = int(stat.Uid)
		foundGid = int(stat.Gid)
	}

	expectedUid, err := utils.GetUserUidFromUsername(t.Data.Owner)
	if err != nil {
		return didSomething, err
	}
	expectedGid, err := utils.GetGroupGidFromName(t.Data.Group)
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
				"[%s] Ownership of the template file (%s) should be (%s:%s) but is (%s:%s)",
				t.String(),
				t.Data.Path,
				t.Data.Owner,
				t.Data.Group,
				username,
				groupName,
			),
		)
		os.Chown(t.Data.Path, expectedUid, expectedGid)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"[%s] Ownership of the template file (%s) updated from (%s:%s) to (%s:%s)",
				t.String(),
				t.Data.Path,
				username,
				groupName,
				t.Data.Owner,
				t.Data.Group,
			),
		)
	}
	return didSomething, nil
}

func (t *TemplateResource) generateTemplate(ctx *models.ResourceContext, templ *template.Template) (string, error) {
	// Build facts map
	factsMap := facts.FactsToMap(*ctx.Facts)

	// Create variables for template
	// 1. First get global variables, from context
	// 2. Copy variables, and override at the same time, with resource scoped variables
	// 3. Set facts in variables
	variables := ctx.Variables
	maps.Copy(variables, t.Data.Variables)
	variables["facts"] = factsMap

	// Build actual template result from variables and template
	var templateBytesResult bytes.Buffer
	err := templ.ExecuteTemplate(&templateBytesResult, t.Title, variables)
	if err != nil {
		return "", err
	}

	return templateBytesResult.String(), nil
}

func (t *TemplateResource) exist() bool {
	if _, err := os.Stat(t.Data.Path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func (t *TemplateResource) create(fileContent string) error {
	file, err := os.Create(t.Data.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get content of file from template, and set it
	file.Write([]byte(fileContent))

	// Change mode of file
	file.Chmod(t.Data.Mode)

	// Set owner and group
	uid, err := utils.GetUserUidFromUsername(t.Data.Owner)
	if err != nil {
		return err
	}
	gid, err := utils.GetGroupGidFromName(t.Data.Group)
	if err != nil {
		return err
	}
	file.Chown(uid, gid)

	return nil
}

func (t *TemplateResource) delete() error {
	err := os.Remove(t.Data.Path)
	return err
}

func (t *TemplateResource) changeContentIfNeeded(expectedContent string) (bool, error) {
	var didSomething bool

	// First we open the file
	file, err := os.Open(t.Data.Path)
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
	io.WriteString(contentMD5, expectedContent)

	fileMD5Checksum := fmt.Sprintf("%x", fileMD5.Sum(nil))
	contentMD5Checksum := fmt.Sprintf("%x", contentMD5.Sum(nil))

	if fileMD5Checksum != contentMD5Checksum {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Checksum for file (%s) should be (%s) but is (%s)",
				t.String(),
				t.Data.Path,
				contentMD5Checksum,
				fileMD5Checksum,
			),
		)
		err := os.WriteFile(t.Data.Path, []byte(expectedContent), t.Data.Mode)
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"[%s] File (%s) content has been updated",
				t.String(),
				t.Data.Path,
			),
		)
		didSomething = true
	}
	return didSomething, nil
}

func (t *TemplateResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	var templateContent string
	var err error

	if t.Data.Source != "" {
		templateContent, err = context.ApiClient.RetrieveTemplate(t.Data.Source, context.Environment, t.Data.roleContext.RoleName)
		if err != nil {
			result.Failed = true
			return result, err
		}
	} else {
		templateContent = t.Data.Content
	}

	templ, err := template.New(t.Title).Parse(templateContent)
	if err != nil {
		result.Failed = true
		return result, err
	}

	templateResult, err := t.generateTemplate(context, templ)
	if err != nil {
		result.Failed = true
		return result, err
	}

	if !t.exist() && t.Present {
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) does not exist, but should", t.String(), t.Data.Path),
		)
		err := t.create(templateResult)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) created", t.String(), t.Data.Path),
		)
		result.Created = true
	} else if t.exist() && !t.Present {
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) exist, but should not", t.String(), t.Data.Path),
		)
		err := t.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) deleted", t.String(), t.Data.Path),
		)
		result.Deleted = true
	}

	if t.exist() && t.Present {
		var err error

		// Check content of the file
		contentHasChanged, err := t.changeContentIfNeeded(templateResult)
		if err != nil {
			result.Failed = true
			return result, err
		}

		// Check permissions of the file
		permissionsHasBeenChanged, err := t.changePermissionsIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		// Check ownership of teh file
		ownershipHasBeenChanged, err := t.changeOwnershipIfNeeded()
		if err != nil {
			result.Failed = true
			return result, err
		}

		if contentHasChanged || permissionsHasBeenChanged || ownershipHasBeenChanged {
			result.Updated = true
		}
	}

	return result, nil
}

func (t *TemplateResource) String() string {
	return fmt.Sprintf("%s / '%s'", t.Type, t.Title)
}

func (t *TemplateResource) When() string {
	return t.WhenCondition
}

func (t *TemplateResource) Register() string {
	return t.RegisterVariable
}

func (t *TemplateResource) Validate() error {
	validationErrors := []models.ValidationError{}

	fieldsThatCannotBeEmpty := [][]string{
		{t.Data.Path, "path"},
		{t.Data.Owner, "owner"},
		{t.Data.Group, "group"},
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

	if t.Data.Source != "" && t.Data.Content != "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "content / source",
				ViolatedRule: "content field and source field are mutually exclusive, you cannot use both.",
			},
		)
	}

	if t.Data.Source != "" && t.Data.roleContext == nil {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "source",
				ViolatedRule: "source field cannot be used outside of roles",
			},
		)
	}

	if len(validationErrors) > 1 {
		return models.ResourceValidationError{
			Type:             t.Type,
			Title:            t.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewTemplateResource(resource *models.Resource, dataField map[string]any, roleContext *models.RoleContext) (*TemplateResource, error) {
	var templateResource TemplateResource

	defaults := map[string]any{
		"owner": "root",
		"group": "root",
		"mode":  0755,
	}

	// Declare data struct
	var templateData TemplateData

	// First we set default values
	err := mapstructure.Decode(defaults, &templateData)
	if err != nil {
		return &templateResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(dataField, &templateData)
	if err != nil {
		return &templateResource, err
	}

	templateResource.Title = resource.Title
	templateResource.Type = resource.Type
	templateResource.Present = *resource.Present
	templateResource.WhenCondition = resource.When
	templateResource.RegisterVariable = resource.Register
	templateResource.Data = templateData
	templateResource.Data.roleContext = roleContext

	// In the case that we didn't have any variables
	if templateResource.Data.Variables == nil {
		templateResource.Data.Variables = map[string]any{}
	}

	return &templateResource, nil
}
