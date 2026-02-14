package inventory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/peeklapp/peekl/pkg/models"
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

type NodeNotFoundError struct {
	NodeName string
}

func (e NodeNotFoundError) Error() string {
	return fmt.Sprintf("The node %s could not be found in the inventory.", e.NodeName)
}

// Load an host from inventory
func LoadNodeFromInventory(codePath string, nodeName string) (*models.NodeInventory, error) {
	var node models.NodeInventory

	// Determine node file path
	nodeFile := filepath.Join(
		codePath,
		"nodes",
		fmt.Sprintf("%s.yml", nodeName),
	)

	// Open file, handle case where it does not exist
	f, err := os.ReadFile(nodeFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &node, NodeNotFoundError{NodeName: nodeName}
		} else {
			return &node, err
		}
	}

	// Load from YAML
	err = yaml.Unmarshal(f, &node)
	if err != nil {
		return &node, err
	}

	return &node, nil
}

type GroupNotFoundError struct {
	GroupName string
}

func (e GroupNotFoundError) Error() string {
	return fmt.Sprintf("The group %s could not be found in the inventory.", e.GroupName)
}

func LoadGroupFromInventory(codePath string, groupName string) (*models.GroupInventory, error) {
	var group models.GroupInventory

	// Determine group file path
	groupFile := filepath.Join(
		codePath,
		"groups",
		fmt.Sprintf("%s.yml", groupName),
	)

	// Open file, handle case where it does not exist
	f, err := os.ReadFile(groupFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &group, GroupNotFoundError{GroupName: groupName}
		} else {
			return &group, err
		}
	}

	// Load from YAML
	err = yaml.Unmarshal(f, &group)
	if err != nil {
		return &group, err
	}

	return &group, nil
}
