package variables

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type VariablesSourceType int

const (
	VariablesSourceNode VariablesSourceType = iota
	VariablesSourceGroup
	VariablesSourceRole
)

func loadVariables(codePath string, entityName string, sourceType VariablesSourceType) (map[string]any, error) {
	variables := map[string]any{}

	var variablesFilesPath string
	switch sourceType {
	case VariablesSourceGroup:
		variablesFilesPath = filepath.Join(codePath, "variables", "groups", entityName, "*.yml")
	case VariablesSourceNode:
		variablesFilesPath = filepath.Join(codePath, "variables", "nodes", entityName, "*.yml")
	case VariablesSourceRole:
		variablesFilesPath = filepath.Join(codePath, "roles", entityName, "variables", "*.yml")
	}

	variablesFiles, err := filepath.Glob(variablesFilesPath)
	if err != nil {
		return variables, err
	}

	for _, variableFile := range variablesFiles {
		rawFile, err := os.ReadFile(variableFile)
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

func LoadGroupVariables(codePath string, groupName string) (map[string]any, error) {
	variables, err := loadVariables(codePath, groupName, VariablesSourceGroup)
	if err != nil {
		return variables, err
	}
	return variables, nil
}

func LoadNodeVariables(codePath string, nodeName string) (map[string]any, error) {
	variables, err := loadVariables(codePath, nodeName, VariablesSourceNode)
	if err != nil {
		return variables, err
	}
	return variables, nil
}

func LoadRoleVariables(codePath string, roleName string) (map[string]any, error) {
	variables, err := loadVariables(codePath, roleName, VariablesSourceRole)
	if err != nil {
		return variables, err
	}
	return variables, nil
}
