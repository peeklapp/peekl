package roles

import (
	"os"
	"testing"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadValidRole(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	role, err := LoadRoleFromCode(root, "nginx")
	if err != nil {
		t.Errorf("Should not have raised an error : %s", err.Error())
	}

	assert.Equal(t, role.Name, "nginx")
	assert.Equal(t, len(role.Entrypoint.Resources), 1)
}

func TestLoadValidRoleWithVars(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	role, _ := LoadRoleFromCode(root, "nginx_vars")
	assert.Equal(t, role.Name, "nginx_vars")
	assert.Equal(t, len(role.Entrypoint.Resources), 1)
	assert.Equal(t, "test", role.Variables["test"])
}

func TestLoadUnknowRole(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	_, err = LoadRoleFromCode(root, "apache")
	assert.Equal(t, "role 'apache' could not be found in the roles folder", err.Error())
}

func TestLoadRoleMissingMain(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	_, err = LoadRoleFromCode(root, "invalid_role")
	assert.ErrorAs(t, err, &models.RoleEntrypointNotFoundError{})
	assert.Equal(t, "could not find any main.hcl file in the 'invalid_role' role", err.Error())
}

func TestLoadRoleMissingInclude(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	_, err = LoadRoleFromCode(root, "missing_include")
	assert.Equal(t, "included file 'test.hcl' could not be found in role 'missing_include'", err.Error())
}

func TestLoadRoleIncludeWithoutHclExtension(t *testing.T) {
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Errorf("Could not open root : %s", err)
	}
	_, err = LoadRoleFromCode(root, "json_include")
	assert.EqualError(t, err, "included file 'testing.json' in role 'json_include' must have a '.hcl' extension")
}
