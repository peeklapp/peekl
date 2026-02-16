package catalog

import (
	"maps"

	"github.com/peeklapp/peekl/pkg/inventory"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/roles"
)

func CompileCatalog(codeDirectory string, nodeName string) ([]models.Resource, []models.Role, []string, map[string]any, error) {
	var resources []models.Resource
	var tags []string
	var rolesToLoad []string

	// Variables
	nodeVars := map[string]any{}
	groupsVars := map[string]any{}
	rolesVars := map[string]any{}
	var loadedRoles []models.Role

	// Handle node
	node, err := inventory.LoadNodeFromInventory(codeDirectory, nodeName)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	for _, nodeRes := range node.Resources {
		resources = append(resources, *nodeRes)
	}
	tags = append(tags, node.Tags...)
	nodeVars = node.Variables
	rolesToLoad = append(rolesToLoad, node.Roles...)

	// Handle groups
	for _, group := range node.Groups {
		groupInv, err := inventory.LoadGroupFromInventory(codeDirectory, group)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		for _, groupRes := range groupInv.Resources {
			resources = append(resources, *groupRes)
		}
		tags = append(tags, groupInv.Tags...)
		maps.Copy(groupsVars, groupInv.Variables)
		rolesToLoad = append(rolesToLoad, groupInv.Roles...)
	}

	// Handle roles
	for _, role := range rolesToLoad {
		loadedRole, err := roles.LoadRoleFromCode(codeDirectory, role)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		loadedRoles = append(loadedRoles, *loadedRole)
		maps.Copy(rolesVars, loadedRole.Variables)
	}

	variables := map[string]any{}
	maps.Copy(variables, rolesVars)
	maps.Copy(variables, groupsVars)
	maps.Copy(variables, nodeVars)

	return resources, loadedRoles, tags, variables, nil
}
