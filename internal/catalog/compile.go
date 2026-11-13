package catalog

import (
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"

	"github.com/peeklapp/peekl/internal/inventory"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/roles"
	"github.com/peeklapp/peekl/internal/variables"
)

// roleSet collects the roles of a node, from the node itself and from its
// groups, in declaration order.
type roleSet struct {
	roles   []models.RoleInventory
	sources map[string]string
}

// add registers the roles declared by source (e.g. "node 'web1'").
//
// A role declared twice by the same source is an error. A role inherited
// from several sources is kept once, as long as every source gives it the
// same variables. Conflicting variables are an error, since there is no
// obvious winner.
func (s *roleSet) add(source string, roles []models.RoleInventory) error {
	seenInSource := map[string]bool{}
	for _, role := range roles {
		if seenInSource[role.Name] {
			return fmt.Errorf("role '%s' has been declared twice in %s", role.Name, source)
		}
		seenInSource[role.Name] = true

		previousSource, present := s.sources[role.Name]
		if !present {
			s.sources[role.Name] = source
			s.roles = append(s.roles, role)
			continue
		}

		idx := slices.IndexFunc(s.roles, func(r models.RoleInventory) bool { return r.Name == role.Name })
		if !reflect.DeepEqual(s.roles[idx].Variables, role.Variables) {
			return fmt.Errorf(
				"role '%s' is declared by both %s and %s with different variables",
				role.Name,
				previousSource,
				source,
			)
		}
	}
	return nil
}

func CompileCatalog(codeRoot *os.Root, nodeName string) ([]models.Role, map[string]any, []string, error) {
	rolesToLoad := roleSet{sources: map[string]string{}}
	tagsFromInventory := []string{}
	nodeVariables := map[string]any{}
	groupVariables := map[string]any{}

	nodes, groups, err := inventory.LoadInventory(codeRoot)
	if err != nil {
		return nil, nil, nil, err
	}

	node, present := nodes[nodeName]
	if !present {
		return nil, nil, nil, models.NodeNotFoundError{NodeName: nodeName}
	}

	if err := rolesToLoad.add(fmt.Sprintf("node '%s'", node.Name), node.Roles); err != nil {
		return nil, nil, nil, err
	}

	if err := variables.LoadNodeVariables(codeRoot, node.Name, &nodeVariables); err != nil {
		return nil, nil, nil, err
	}
	tagsFromInventory = append(tagsFromInventory, node.Tags...)

	for _, groupName := range node.Groups {
		group, present := groups[groupName]
		if !present {
			return nil, nil, nil, models.GroupNotFoundError{GroupName: groupName}
		}

		if err := rolesToLoad.add(fmt.Sprintf("group '%s'", groupName), group.Roles); err != nil {
			return nil, nil, nil, err
		}

		if err := variables.LoadGroupVariables(codeRoot, groupName, &groupVariables); err != nil {
			return nil, nil, nil, err
		}
		tagsFromInventory = append(tagsFromInventory, group.Tags...)
	}

	var loadedRoles []models.Role
	for _, roleToLoad := range rolesToLoad.roles {
		role, err := roles.LoadRoleFromCode(codeRoot, roleToLoad.Name)
		if err != nil {
			return nil, nil, nil, err
		}
		maps.Copy(role.Variables, roleToLoad.Variables)
		loadedRoles = append(loadedRoles, *role)
	}

	// Overwrite node variables on top of group variables
	maps.Copy(groupVariables, nodeVariables)

	return loadedRoles, groupVariables, tagsFromInventory, nil
}
