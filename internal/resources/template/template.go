package template

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"syscall"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/internal/facts"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/resources"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/sirupsen/logrus"
)

type TemplateParameters struct {
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
	Parameters TemplateParameters
}

func (t *TemplateResource) changePermissionsIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(t.Parameters.Path)
	if err != nil {
		return didSomething, err
	}

	// Update file permissions if needed
	if stat.Mode() != t.Parameters.Mode {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Mode for the template file (%s) should be (%s) but is (%s)",
				t.String(),
				t.Parameters.Path,
				t.Parameters.Mode,
				stat.Mode(),
			),
		)
		err = os.Chmod(t.Parameters.Path, t.Parameters.Mode)
		if err != nil {
			return didSomething, err
		}
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"[%s] Mode for the template file (%s) has been updated from (%s) to (%s)",
				t.String(),
				t.Parameters.Path,
				stat.Mode(),
				t.Parameters.Mode,
			),
		)
	}
	return didSomething, nil
}

func (t *TemplateResource) changeOwnershipIfNeeded() (bool, error) {
	var didSomething bool

	// Get stat for the file
	stat, err := os.Stat(t.Parameters.Path)
	if err != nil {
		return didSomething, err
	}

	var foundUid int
	var foundGid int

	if stat, ok := stat.Sys().(*syscall.Stat_t); ok {
		foundUid = int(stat.Uid)
		foundGid = int(stat.Gid)
	}

	expectedUid, err := utils.GetUserUidFromUsername(t.Parameters.Owner)
	if err != nil {
		return didSomething, err
	}
	expectedGid, err := utils.GetGroupGidFromName(t.Parameters.Group)
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
				t.Parameters.Path,
				t.Parameters.Owner,
				t.Parameters.Group,
				username,
				groupName,
			),
		)
		err = os.Chown(t.Parameters.Path, expectedUid, expectedGid)
		if err != nil {
			return didSomething, err
		}
		didSomething = true
		logrus.Info(
			fmt.Sprintf(
				"[%s] Ownership of the template file (%s) updated from (%s:%s) to (%s:%s)",
				t.String(),
				t.Parameters.Path,
				username,
				groupName,
				t.Parameters.Owner,
				t.Parameters.Group,
			),
		)
	}
	return didSomething, nil
}

func (t *TemplateResource) generateTemplate(ctx *models.ResourceContext, templ *template.Template) (string, error) {
	// Build facts map
	factsMap, err := facts.FactsToMap(*ctx.Facts)
	if err != nil {
		return "", err
	}

	// Create variables for template
	// 1. First get global variables, from context
	// 2. Copy variables, and override at the same time, with resource scoped variables
	// 3. Set facts in variables
	variables := ctx.Variables
	maps.Copy(variables, t.Parameters.Variables)
	variables["facts"] = factsMap

	// Build actual template result from variables and template
	var templateBytesResult bytes.Buffer
	err = templ.ExecuteTemplate(&templateBytesResult, t.Title, variables)
	if err != nil {
		return "", err
	}

	return templateBytesResult.String(), nil
}

func (t *TemplateResource) exist() bool {
	return utils.FileExist(t.Parameters.Path, nil)
}

func (t *TemplateResource) create(fileContent string) error {
	file, err := os.Create(t.Parameters.Path)
	if err != nil {
		return err
	}
	defer utils.CloseWithoutError(file)

	// Get content of file from template, and set it
	_, err = file.Write([]byte(fileContent))
	if err != nil {
		return err
	}

	// Change mode of file
	err = file.Chmod(t.Parameters.Mode)
	if err != nil {
		return err
	}

	// Set owner and group
	uid, err := utils.GetUserUidFromUsername(t.Parameters.Owner)
	if err != nil {
		return err
	}
	gid, err := utils.GetGroupGidFromName(t.Parameters.Group)
	if err != nil {
		return err
	}

	err = file.Chown(uid, gid)
	if err != nil {
		return err
	}

	return err
}

func (t *TemplateResource) delete() error {
	err := os.Remove(t.Parameters.Path)
	return err
}

func (t *TemplateResource) changeContentIfNeeded(expectedContent string) (bool, error) {
	var didSomething bool

	// First we open the file
	file, err := os.Open(t.Parameters.Path)
	if err != nil {
		return didSomething, err
	}
	defer utils.CloseWithoutError(file)

	// Then we create MD5 object of file
	fileMD5 := md5.New()
	if _, err := io.Copy(fileMD5, file); err != nil {
		return didSomething, err
	}

	// Then we create MD5 object of content
	contentMD5 := md5.New()
	_, err = io.WriteString(contentMD5, expectedContent)
	if err != nil {
		return didSomething, err
	}

	fileMD5Checksum := fmt.Sprintf("%x", fileMD5.Sum(nil))
	contentMD5Checksum := fmt.Sprintf("%x", contentMD5.Sum(nil))

	if fileMD5Checksum != contentMD5Checksum {
		logrus.Info(
			fmt.Sprintf(
				"[%s] Checksum for file (%s) should be (%s) but is (%s)",
				t.String(),
				t.Parameters.Path,
				contentMD5Checksum,
				fileMD5Checksum,
			),
		)
		err := os.WriteFile(t.Parameters.Path, []byte(expectedContent), t.Parameters.Mode)
		if err != nil {
			return didSomething, err
		}
		logrus.Info(
			fmt.Sprintf(
				"[%s] File (%s) content has been updated",
				t.String(),
				t.Parameters.Path,
			),
		)
		didSomething = true
	}

	return didSomething, err
}

func (t *TemplateResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult

	var templateContent string
	var err error

	if t.Parameters.Source != "" {
		fullTemplatePath := filepath.Join(context.CodePath, "roles", t.Parameters.roleContext.RoleName, "templates", t.Parameters.Source)
		if !utils.FileExist(fullTemplatePath, nil) {
			result.Failed = true
			return result, fmt.Errorf("[%s] The template doesn't exist in role : %s", t.String(), fullTemplatePath)
		}
		rawTemplateContent, err := os.ReadFile(fullTemplatePath)
		if err != nil {
			result.Failed = true
			return result, err
		}
		templateContent = string(rawTemplateContent)
	} else {
		templateContent = t.Parameters.Content
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
			fmt.Sprintf("[%s] Template file (%s) does not exist, but should", t.String(), t.Parameters.Path),
		)
		err := t.create(templateResult)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) created", t.String(), t.Parameters.Path),
		)
		result.Created = true
	} else if t.exist() && !t.Present {
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) exist, but should not", t.String(), t.Parameters.Path),
		)
		err := t.delete()
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("[%s] Template file (%s) deleted", t.String(), t.Parameters.Path),
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
		{t.Parameters.Path, "path"},
		{t.Parameters.Owner, "owner"},
		{t.Parameters.Group, "group"},
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

	if t.Parameters.Source != "" && t.Parameters.Content != "" {
		validationErrors = append(
			validationErrors,
			models.ValidationError{
				FieldName:    "content / source",
				ViolatedRule: "content field and source field are mutually exclusive, you cannot use both.",
			},
		)
	}

	if t.Parameters.Source != "" && t.Parameters.roleContext == nil {
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

func NewTemplateResource(resource *models.Resource, parametersField map[string]any, roleContext *models.RoleContext) (*TemplateResource, error) {
	var templateResource TemplateResource

	defaults := map[string]any{
		"owner": "root",
		"group": "root",
		"mode":  0755,
	}

	// Declare parameters struct
	var templateParameters TemplateParameters

	// First we set default values
	err := mapstructure.Decode(defaults, &templateParameters)
	if err != nil {
		return &templateResource, err
	}

	// Then we override with actual values
	err = mapstructure.Decode(parametersField, &templateParameters)
	if err != nil {
		return &templateResource, err
	}

	templateResource.Title = resource.Title
	templateResource.Type = resource.Type
	templateResource.Present = *resource.Present
	templateResource.WhenCondition = resource.When
	templateResource.RegisterVariable = resource.Register
	templateResource.Parameters = templateParameters
	templateResource.Parameters.roleContext = roleContext

	// In the case that we didn't have any variables
	if templateResource.Parameters.Variables == nil {
		templateResource.Parameters.Variables = map[string]any{}
	}

	return &templateResource, nil
}
