package roles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
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
	if err := DoesRoleExist(codeRoot, roleName); err != nil {
		return nil, err
	}

	var role models.Role
	role.Name = roleName

	entrypointFilePath := filepath.Join("roles", roleName, "main.hcl")
	entrypointFile, err := codeRoot.ReadFile(entrypointFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, models.RoleEntrypointNotFoundError{RoleName: roleName}
		}
		return nil, err
	}
	if err := hclsimple.Decode(entrypointFilePath, entrypointFile, nil, &role.Entrypoint); err != nil {
		return nil, err
	}
	for i := range role.Entrypoint.Resources {
		if err := role.Entrypoint.Resources[i].Resolve(); err != nil {
			return nil, err
		}
	}

	role.Included = map[string]models.RoleInclude{}
	for _, includedFile := range role.Entrypoint.Includes {
		currIncludeFilePath := filepath.Join("roles", roleName, includedFile.Path)
		if filepath.Ext(includedFile.Path) != ".hcl" {
			return nil, fmt.Errorf("included file '%s' in role '%s' must have a '.hcl' extension", includedFile.Path, roleName)
		}
		currIncludeFileRaw, err := codeRoot.ReadFile(currIncludeFilePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("included file '%s' could not be found in role '%s'", includedFile.Path, roleName)
			} else {
				return nil, err
			}
		}

		var currIncludedFile models.RoleInclude
		if err := hclsimple.Decode(currIncludeFilePath, currIncludeFileRaw, nil, &currIncludedFile); err != nil {
			return nil, err
		}
		for i := range currIncludedFile.Resources {
			if err := currIncludedFile.Resources[i].Resolve(); err != nil {
				return nil, err
			}
		}
		role.Included[includedFile.Name] = currIncludedFile
	}

	role.Variables = map[string]any{}
	if err := variables.LoadRoleVariables(codeRoot, roleName, &role.Variables); err != nil {
		return nil, err
	}

	return &role, nil
}
