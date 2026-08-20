package roles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	yaml "github.com/goccy/go-yaml"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/peeklapp/peekl/internal/variables"
)

func DoesRoleExist(codeRoot *os.Root, roleName string) error {
	if !utils.FileExist(filepath.Join("roles", roleName), codeRoot) {
		return models.RoleNotFoundError{RoleName: roleName}
	}
	return nil
}

func LoadRoleFromCode(codeRoot *os.Root, roleName string) (*models.Role, error) {
	var role models.Role

	role.Name = roleName
	role.IncludedResources = map[string]models.IncludedResources{}

	err := DoesRoleExist(codeRoot, roleName)
	if err != nil {
		return &role, err
	}

	// Open root of role
	rolePath := filepath.Join("roles", roleName)
	roleRoot, err := codeRoot.OpenRoot(rolePath)
	if err != nil {
		return &role, fmt.Errorf("could not open root inside role when loading role %s", roleName)
	}

	// Open main.yml file, handle error if it does not exist
	mainFile, err := roleRoot.ReadFile("main.yml")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &role, fmt.Errorf("could not find any main.yml file in the %s role", roleName)
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

	if len(roleMain.DependsOn) == 0 {
		role.DependsOn = []string{}
	} else {
		role.DependsOn = roleMain.DependsOn
	}

	// For each include in roleMain, include resources
	if len(roleMain.Includes) > 0 {
		// For each extra file, process
		for _, extraFile := range roleMain.Includes {
			// Open extra file, handle error if it does not exist
			rawExtraFile, err := roleRoot.ReadFile(fmt.Sprintf("%s.yml", extraFile.Name))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return &role, fmt.Errorf("the include `%s` in role `%s` could not be found", extraFile.Name, roleName)
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

	role.Variables, err = variables.LoadRoleVariables(codeRoot, roleName)
	if err != nil {
		return &role, err
	}

	return &role, nil
}

func RoleNameIsValid(roleName string) bool {
	r, _ := regexp.Compile("^[A-Za-z0-9_]+$")
	return r.MatchString(roleName)
}
