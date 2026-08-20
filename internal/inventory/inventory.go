package inventory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/variables"
)

// Inventory
//
// The inventory is composed of two different
// main directory: the `nodes` directory and
// the `groups` directory.
//
// A node declaration contains information about
// the node. Such as the groups he's a member of,
// the roles that should apply to it.
//
// A group otherwise correspond to a list of
// roles that are getting applied only. A group
// is not aware of his members. Only a node knows
// of which group he's a member of.
//
// eg:
//    code/
//      nodes/
//        node-1.yml
//        ...
//      groups/
//        web.yml
//        ...
//      roles/
//        nginx/
//        ...

// Load an host from inventory
func LoadNodeFromInventory(codeRoot *os.Root, nodeName string) (*models.NodeInventory, error) {
	var node models.NodeInventory

	// Determine node file path
	nodeFile := filepath.Join(
		"inventory",
		"nodes",
		fmt.Sprintf("%s.yml", nodeName),
	)

	// Open file, handle case where it does not exist
	f, err := codeRoot.ReadFile(nodeFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &node, models.NodeNotFoundError{NodeName: nodeName}
		} else {
			return &node, err
		}
	}

	// Load from YAML
	err = yaml.Unmarshal(f, &node)
	if err != nil {
		return &node, err
	}

	// Load variables
	node.Variables, err = variables.LoadNodeVariables(codeRoot, nodeName)
	if err != nil {
		return &node, err
	}

	return &node, nil
}

func LoadGroupFromInventory(codeRoot *os.Root, groupName string) (*models.GroupInventory, error) {
	var group models.GroupInventory

	// Determine group file path
	groupFile := filepath.Join(
		"inventory",
		"groups",
		fmt.Sprintf("%s.yml", groupName),
	)

	// Open file, handle case where it does not exist
	f, err := codeRoot.ReadFile(groupFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &group, models.GroupNotFoundError{GroupName: groupName}
		} else {
			return &group, err
		}
	}

	// Load from YAML
	err = yaml.Unmarshal(f, &group)
	if err != nil {
		return &group, err
	}

	// Load variables
	group.Variables, err = variables.LoadGroupVariables(codeRoot, groupName)
	if err != nil {
		return &group, err
	}

	return &group, nil
}
