package inventory

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
	"github.com/peeklapp/peekl/internal/variables"
)

// TODO: ADD PROPER ROOT TO PREVENT FILE ESCAPE
func LoadNodeFromInventory(codePath string, nodeName string) (*models.NodeInventory, error) {
	var node models.NodeInventory

	nodeFile := filepath.Join(
		codePath,
		"inventory",
		"nodes",
		fmt.Sprintf("%s.hcl", nodeName),
	)

	if !utils.FileExist(nodeFile) {
		return &node, models.NodeNotFoundError{NodeName: nodeName}
	}

	err := hclsimple.DecodeFile(nodeFile, nil, &node)
	if err != nil {
		return &node, err
	}

	node.Variables, err = variables.LoadNodeVariables(codePath, nodeName)
	if err != nil {
		return &node, err
	}

	return &node, nil
}

func LoadGroupFromInventory(codePath string, groupName string) (*models.GroupInventory, error) {
	var group models.GroupInventory

	// Determine group file path
	groupFile := filepath.Join(
		codePath,
		"inventory",
		"groups",
		fmt.Sprintf("%s.yml", groupName),
	)

	if !utils.FileExist(groupFile) {
		return &group, models.GroupNotFoundError{GroupName: groupName}
	}

	err := hclsimple.DecodeFile(groupFile, nil, &group)
	if err != nil {
		return &group, err
	}

	group.Variables, err = variables.LoadGroupVariables(codePath, groupName)
	if err != nil {
		return &group, err
	}

	return &group, nil
}
