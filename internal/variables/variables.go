package variables

import (
	"io/fs"
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
)

type VariablesSourceType int

const (
	VariablesSourceNode VariablesSourceType = iota
	VariablesSourceGroup
	VariablesSourceRole
)

func loadVariables(codeRoot *os.Root, entityName string, sourceType VariablesSourceType) (map[string]any, error) {
	variables := map[string]any{}

	var variablesFilesPath string
	switch sourceType {
	case VariablesSourceGroup:
		variablesFilesPath = filepath.Join("variables", "groups", entityName, "*.yml")
	case VariablesSourceNode:
		variablesFilesPath = filepath.Join("variables", "nodes", entityName, "*.yml")
	case VariablesSourceRole:
		variablesFilesPath = filepath.Join("roles", entityName, "variables", "*.yml")
	}

	variablesFiles, err := fs.Glob(codeRoot.FS(), variablesFilesPath)
	if err != nil {
		return variables, err
	}

	for _, variableFile := range variablesFiles {
		rawFile, err := codeRoot.ReadFile(variableFile)
		if err != nil {
			return variables, err
		}
		err = yaml.Unmarshal(rawFile, &variables)
		if err != nil {
			return variables, err
		}
	}

	return variables, nil
}

func LoadGroupVariables(codeRoot *os.Root, groupName string) (map[string]any, error) {
	variables, err := loadVariables(codeRoot, groupName, VariablesSourceGroup)
	if err != nil {
		return variables, err
	}
	return variables, nil
}

func LoadNodeVariables(codeRoot *os.Root, nodeName string) (map[string]any, error) {
	variables, err := loadVariables(codeRoot, nodeName, VariablesSourceNode)
	if err != nil {
		return variables, err
	}
	return variables, nil
}

func LoadRoleVariables(codeRoot *os.Root, roleName string) (map[string]any, error) {
	variables, err := loadVariables(codeRoot, roleName, VariablesSourceRole)
	if err != nil {
		return variables, err
	}
	return variables, nil
}
