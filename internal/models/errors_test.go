package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleNotFoundError(t *testing.T) {
	err := RoleNotFoundError{
		RoleName: "my_role",
	}
	assert.Equal(t, "role 'my_role' could not be found in the roles folder", err.Error())
}

func TestRoleEntrypointNotFoundError(t *testing.T) {
	err := RoleEntrypointNotFoundError{
		RoleName: "my_role",
	}
	assert.Equal(t, "could not find any main.hcl file in the 'my_role' role", err.Error())
}

func TestNodeNotFoundError(t *testing.T) {
	err := NodeNotFoundError{
		NodeName: "my_node",
	}
	assert.Equal(t, "node 'my_node' could not be found in the inventory", err.Error())
}

func TestGroupNotFoundError(t *testing.T) {
	err := GroupNotFoundError{
		GroupName: "my_group",
	}
	assert.Equal(t, "group 'my_group' could not be found in the inventory", err.Error())
}
