package roles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/variables"
)

func LoadRoleFromCode(codePath string, roleName string) (*models.Role, error) {
	var role models.Role

	// Set role name
	role.Name = roleName

	// Make sure template map is initilized
	role.Templates = map[string]string{}

	// Make sure to initialize map
	role.IncludedResources = map[string]models.IncludedResources{}

	// Make sure role exist locally
	rolePath := filepath.Join(codePath, "roles", roleName)
	if _, err := os.Stat(rolePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &role, models.RoleNotFoundError{RoleName: roleName}
		} else {
			return &role, err
		}
	}

	// Open main.yml file, handle error if it does not exist
	mainFile, err := os.ReadFile(filepath.Join(rolePath, "main.yml"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &role, fmt.Errorf("Could not find any main.yml file in the %s role.", roleName)
		} else {
			return &role, err
		}
	}

	// Load main file into struct
	var roleMain models.RoleMain
	err = yaml.Unmarshal(mainFile, &roleMain)
	if err != nil {
		return &role, err
	}

	// Append resources of role main to role resources
	role.Resources = roleMain.Resources

	// For each include in roleMain, include resources
	if len(roleMain.Includes) > 0 {
		// For each extra file, process
		for _, extraFile := range roleMain.Includes {
			// Open extra file, handle error if it does not exist
			rawExtraFile, err := os.ReadFile(filepath.Join(rolePath, fmt.Sprintf("%s.yml", extraFile.Name)))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return &role, fmt.Errorf("The include `%s` in role `%s` could not be found.", extraFile.Name, roleName)
				} else {
					return &role, err
				}
			}

			// Load file resources
			var resources []models.Resource
			err = yaml.Unmarshal(rawExtraFile, &resources)
			if err != nil {
				return &role, err
			}

			// Append to Role conditional resources
			role.IncludedResources[extraFile.Name] = models.IncludedResources{
				Resources: resources,
			}
		}
	}

	// Make sure templates directory actually exist
	var templatesDirExist bool
	templatePath := filepath.Join(rolePath, "templates")
	_, err = os.Stat(templatePath)
	if err == nil {
		templatesDirExist = true
	} else {
		if errors.Is(err, os.ErrNotExist) {
			templatesDirExist = false
		} else {
			return &role, err
		}
	}

	if templatesDirExist {
		// Find all template files
		templatePathGlob := filepath.Join(templatePath, "*.tmpl")
		templateFiles, err := filepath.Glob(templatePathGlob)
		if err != nil {
			return &role, err
		}

		// Load all template file
		for _, templateFile := range templateFiles {
			rawTemplateFile, err := os.ReadFile(templateFile)
			if err != nil {
				return &role, err
			}
			role.Templates[strings.Split(filepath.Base(templateFile), ".tmpl")[0]] = string(rawTemplateFile)
		}
	}

	role.Variables, err = variables.LoadRoleVariables(codePath, roleName)
	if err != nil {
		return &role, err
	}

	return &role, nil
}
