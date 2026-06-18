package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoleNotFoundError(t *testing.T) {
	err := RoleNotFoundError{
		RoleName: "my_role",
	}
	assert.Equal(t, "The role my_role could not be found in the roles folder", err.Error())
}

func TestNodeNotFoundError(t *testing.T) {
	err := NodeNotFoundError{
		NodeName: "my_node",
	}
	assert.Equal(t, "The node my_node could not be found in the inventory", err.Error())
}

func TestGroupNotFoundError(t *testing.T) {
	err := GroupNotFoundError{
		GroupName: "my_group",
	}
	assert.Equal(t, "The group my_group could not be found in the inventory", err.Error())
}

func TestEnvironmentNotFoundError(t *testing.T) {
	err := EnvironmentNotFoundError{
		Environment: "my_environment",
	}
	assert.Equal(t, "The environment my_environment could not be found in the code folder", err.Error())
}
