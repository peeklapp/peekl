package template

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"syscall"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources"
	"github.com/peeklapp/peekl/pkg/utils"
	"github.com/sirupsen/logrus"
)

type TemplateData struct {
	Name           string         `mapstructure:"name"`
	Path           string         `mapstructure:"path"`
	Owner          string         `mapstructure:"owner"`
	Group          string         `mapstrucutre:"group"`
	Mode           fs.FileMode    `mapstructure:"mode"`
	Variables      map[string]any `mapstructure:"variables"`
	rawTemplateDir string
	rawTemplate    string
	loadedTemplate template.Template
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
				"Mode for the template file (%s) should be (%s) but is (%s)",
				t.Data.Path,
				t.Data.Mode,
				stat.Mode(),
			),
		)
		os.Chmod(t.Data.Path, t.Data.Mode)
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"Mode for the template file (%s) has been updated from (%s) to (%s)",
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
				"Ownership of the template file (%s) should be (%s:%s) but is (%s:%s)",
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
				"Ownership of the template file (%s) updated from (%s:%s) to (%s:%s)",
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

func (t *TemplateResource) generateTemplate(ctx *models.ResourceContext) (string, error) {
	// Build facts map
	jsonFacts, err := json.Marshal(ctx.Facts)
	if err != nil {
		return "", err
	}
	var factsMap map[string]any
	err = json.Unmarshal(jsonFacts, &factsMap)
	if err != nil {
		return "", err
	}

	// Create variables for template
	// 1. First get global variables, from context
	// 2. Copy variables, and override at the same time, with resource scoped variables
	// 3. Set facts in variables
	variables := ctx.Variables
	maps.Copy(variables, t.Data.Variables)
	variables["facts"] = factsMap

	// Build actual template result from variables and template
	var templateBytesResult bytes.Buffer
	err = t.Data.loadedTemplate.ExecuteTemplate(&templateBytesResult, fmt.Sprintf("%s", t.Data.Name), variables)
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
				"Checksum for file (%s) should be (%s) but is (%s)",
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
				"File (%s) content has been updated",
				t.Data.Path,
			),
		)
		didSomething = true
	}
	return didSomething, nil
}

func (t *TemplateResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	templateResult, err := t.generateTemplate(context)
	if err != nil {
		result.Failed = true
		return result, err
	}

	if !t.exist() && t.Present {
		logrus.Info(
			fmt.Sprintf("Template file (%s) does not exist, but should", t.Data.Path),
		)
		err := t.create(templateResult)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Temaplte file (%s) created", t.Data.Path),
		)
		result.Created = true
	} else if t.exist() && !t.Present {
		logrus.Info(
			fmt.Sprintf("Template file (%s) exist, but should not", t.Data.Path),
		)
		err := t.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Template file (%s) deleted", t.Data.Path),
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
		{t.Data.Name, "name"},
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

	if len(validationErrors) > 1 {
		return models.ResourceValidationError{
			Type:             t.Type,
			Title:            t.Title,
			ValidationErrors: validationErrors,
		}
	}

	return nil
}

func NewTemplateResource(resource *models.Resource, dataField map[string]any, templates map[string]string) (*TemplateResource, error) {
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

	// Load raw template
	currTemp := templates[templateData.Name]
	rawTemplate, err := template.New(templateResource.Data.Name).Parse(currTemp)
	if err != nil {
		return &templateResource, err
	}
	templateResource.Data.loadedTemplate = *rawTemplate

	// In the case that we didn't have any variables
	if templateResource.Data.Variables == nil {
		templateResource.Data.Variables = map[string]any{}
	}

	return &templateResource, nil
}
