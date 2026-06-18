package roles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadValidRole(t *testing.T) {
	role, err := LoadRoleFromCode("testdata", "nginx")
	if err != nil {
		t.Errorf("Should not have raised an error : %s", err.Error())
	}

	assert.Equal(t, role.Name, "nginx")
	assert.Equal(t, len(role.IncludedResources), 1)
	assert.Equal(t, len(role.Resources), 1)
}

func TestLoadValidRoleWithVars(t *testing.T) {
	role, err := LoadRoleFromCode("testdata", "nginx_vars")
	if err != nil {
		t.Errorf("Should not have raised an error : %s", err.Error())
	}

	assert.Equal(t, role.Name, "nginx_vars")
	assert.Equal(t, len(role.IncludedResources), 1)
	assert.Equal(t, len(role.Resources), 1)
	assert.Equal(t, role.Variables["test"], "test")
}

func TestLoadUnknowRole(t *testing.T) {
	_, err := LoadRoleFromCode("testdata", "apache")
	assert.Equal(t, err.Error(), "The role apache could not be found in the roles folder")
}

func TestLoadRoleMissingMain(t *testing.T) {
	_, err := LoadRoleFromCode("testdata", "invalid_role")
	assert.Equal(t, err.Error(), "Could not find any main.yml file in the invalid_role role.")
}

func TestLoadRoleMissingInclude(t *testing.T) {
	_, err := LoadRoleFromCode("testdata", "missing_include")
	assert.Equal(t, err.Error(), "The include `test` in role `missing_include` could not be found.")
}
