package variables

import (
	"fmt"
	"maps"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/zclconf/go-cty/cty"
)

type VariablesSourceType int

const (
	VariablesSourceNode VariablesSourceType = iota
	VariablesSourceGroup
	VariablesSourceRole
)

func ctyToGo(v cty.Value) (any, error) {
	// A cty.IsNull should be converted to nil
	if v.IsNull() {
		return nil, nil
	}

	// Handle simple type such as string, bool, and number
	switch v.Type() {
	case cty.String:
		return v.AsString(), nil
	case cty.Bool:
		return v.True(), nil
	case cty.Number:
		f, _ := v.AsBigFloat().Float64()
		return f, nil
	}

	// Handle tuple/list case (recursive)
	if v.Type().IsTupleType() || v.Type().IsListType() {
		var out []any
		for it := v.ElementIterator(); it.Next(); {
			_, ev := it.Element()
			converted, err := ctyToGo(ev)
			if err != nil {
				return nil, err
			}
			out = append(out, converted)
		}
		return out, nil
	}

	// Handle dict
	if v.Type().IsObjectType() || v.Type().IsMapType() {
		out := make(map[string]any)
		for it := v.ElementIterator(); it.Next(); {
			k, ev := it.Element()
			converted, err := ctyToGo(ev)
			if err != nil {
				return nil, err
			}
			out[k.AsString()] = converted
		}
		return out, nil
	}

	return nil, fmt.Errorf("Value could not be identified : %+s", v)
}

func convertRawVariables(toBeConverted map[string]cty.Value) (map[string]any, error) {
	var convertedRawVariables = map[string]any{}
	for key, value := range toBeConverted {
		converted, err := ctyToGo(value)
		if err != nil {
			return convertedRawVariables, err
		}
		convertedRawVariables[key] = converted
	}
	return convertedRawVariables, nil
}

// TODO: CREATE CHROOT TO PREVENT PATH ESCAPING
func loadVariables(codePath string, entityName string, sourceType VariablesSourceType) (map[string]any, error) {
	variables := map[string]any{}

	var variablesFilesPath string
	switch sourceType {
	case VariablesSourceGroup:
		variablesFilesPath = filepath.Join(codePath, "variables", "groups", entityName, "*.hcl")
	case VariablesSourceNode:
		variablesFilesPath = filepath.Join(codePath, "variables", "nodes", entityName, "*.hcl")
	case VariablesSourceRole:
		variablesFilesPath = filepath.Join(codePath, "roles", entityName, "variables", "*.hcl")
	}

	variablesFiles, err := filepath.Glob(variablesFilesPath)
	if err != nil {
		return variables, err
	}

	for _, variableFile := range variablesFiles {
		var rawVariables map[string]cty.Value
		err = hclsimple.DecodeFile(variableFile, nil, &rawVariables)
		if err != nil {
			return variables, err
		}

		convertedRawVariables, err := convertRawVariables(rawVariables)
		if err != nil {
			return variables, err
		}
		maps.Copy(variables, convertedRawVariables)
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
