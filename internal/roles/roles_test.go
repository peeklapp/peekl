package roles

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadValidRole(t *testing.T) {
	root, _ := os.OpenRoot("testdata")
	role, err := LoadRoleFromCode(root, "nginx")
	if err != nil {
		t.Errorf("Should not have raised an error : %s", err.Error())
	}

	assert.Equal(t, role.Name, "nginx")
	assert.Equal(t, len(role.IncludedResources), 1)
	assert.Equal(t, len(role.Resources), 1)
}

func TestLoadValidRoleWithVars(t *testing.T) {
	root, _ := os.OpenRoot("testdata")
	role, _ := LoadRoleFromCode(root, "nginx_vars")
	assert.Equal(t, role.Name, "nginx_vars")
	assert.Equal(t, len(role.IncludedResources), 1)
	assert.Equal(t, len(role.Resources), 1)
	assert.Equal(t, role.Variables["test"], "test")
}

func TestLoadUnknowRole(t *testing.T) {
	root, _ := os.OpenRoot("testdata")
	_, err := LoadRoleFromCode(root, "apache")
	assert.Equal(t, err.Error(), "The role apache could not be found in the roles folder")
}

func TestLoadRoleMissingMain(t *testing.T) {
	root, _ := os.OpenRoot("testdata")
	_, err := LoadRoleFromCode(root, "invalid_role")
	assert.Equal(t, err.Error(), "could not find any main.yml file in the invalid_role role")
}

func TestLoadRoleMissingInclude(t *testing.T) {
	root, _ := os.OpenRoot("testdata")
	_, err := LoadRoleFromCode(root, "missing_include")
	assert.Equal(t, err.Error(), "the include `test` in role `missing_include` could not be found")
}
