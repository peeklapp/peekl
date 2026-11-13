package variables

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/zclconf/go-cty/cty"
)

type VariablesSourceType int

const (
	VariablesSourceNode VariablesSourceType = iota
	VariablesSourceGroup
	VariablesSourceRole
)

func loadVariables(codeRoot *os.Root, entityName string, sourceType VariablesSourceType, variablesStore *map[string]any) error {
	var variablesFilesPath string
	switch sourceType {
	case VariablesSourceGroup:
		variablesFilesPath = filepath.Join("variables", "groups", entityName, "*.hcl")
	case VariablesSourceNode:
		variablesFilesPath = filepath.Join("variables", "nodes", entityName, "*.hcl")
	case VariablesSourceRole:
		variablesFilesPath = filepath.Join("roles", entityName, "variables", "*.hcl")
	}

	variablesFiles, err := fs.Glob(codeRoot.FS(), variablesFilesPath)
	if err != nil {
		return err
	}

	tmpVariables := map[string]cty.Value{}
	for _, variableFile := range variablesFiles {
		rawFile, err := codeRoot.ReadFile(variableFile)
		if err != nil {
			return err
		}
		fileVars := map[string]cty.Value{}
		if err := hclsimple.Decode(variableFile, rawFile, nil, &fileVars); err != nil {
			return err
		}
		for k, v := range fileVars {
			tmpVariables[k] = v
		}
	}

	variables, err := models.CtyToMap(cty.ObjectVal(tmpVariables))
	if err != nil {
		return err
	}
	maps.Copy(*variablesStore, variables)

	return nil
}

func LoadGroupVariables(codeRoot *os.Root, groupName string, variablesStore *map[string]any) error {
	return loadVariables(codeRoot, groupName, VariablesSourceGroup, variablesStore)
}

func LoadNodeVariables(codeRoot *os.Root, nodeName string, variablesStore *map[string]any) error {
	return loadVariables(codeRoot, nodeName, VariablesSourceNode, variablesStore)
}

func LoadRoleVariables(codeRoot *os.Root, roleName string, variablesStore *map[string]any) error {
	return loadVariables(codeRoot, roleName, VariablesSourceRole, variablesStore)
}
