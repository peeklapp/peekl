package catalog

import (
	"os"
	"testing"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadNodeFromCode(t *testing.T) {
	root, err := os.OpenRoot("testdata/valid_code")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	loadedRoles, variables, tags, err := CompileCatalog(root, "dummy")
	if err != nil {
		t.Errorf("Could not load catalog from code : %s", err.Error())
	}
	assert.Equal(t, 2, len(loadedRoles))
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, 0, len(variables))
}

func TestInheritedRoleIsDeduplicated(t *testing.T) {
	root, err := os.OpenRoot("testdata/duplicate_role")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	loadedRoles, _, _, err := CompileCatalog(root, "dummy")
	if err != nil {
		t.Fatalf("Should not have returned an error : %s", err.Error())
	}
	var names []string
	for _, r := range loadedRoles {
		names = append(names, r.Name)
	}
	assert.Equal(t, []string{"test", "nginx"}, names)
}

func TestConflictingRoleVariables(t *testing.T) {
	root, err := os.OpenRoot("testdata/conflicting_role_variables")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	_, _, _, err = CompileCatalog(root, "dummy")
	if err == nil {
		t.Fatalf("Should have returned an error since the role variables conflict")
	}
	assert.Equal(t, "role 'nginx' is declared by both node 'dummy' and group 'web' with different variables", err.Error())
}

func TestLoadNodeMissingGroup(t *testing.T) {
	root, err := os.OpenRoot("testdata/missing_group")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	_, _, _, err = CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a group is missing")
	}
	assert.ErrorAs(t, err, &models.GroupNotFoundError{})
	assert.Equal(t, "group 'web' could not be found in the inventory", err.Error())
}

func TestLoadNodeMissingRole(t *testing.T) {
	root, err := os.OpenRoot("testdata/missing_role")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	_, _, _, err = CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.Equal(t, "role 'test' could not be found in the roles folder", err.Error())
}

func TestLoadNodeMissingNode(t *testing.T) {
	root, err := os.OpenRoot("testdata/missing_node")
	if err != nil {
		t.Errorf("Could not have open a root : %s", err.Error())
	}
	_, _, _, err = CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.ErrorAs(t, err, &models.NodeNotFoundError{})
	assert.Equal(t, "node 'dummy' could not be found in the inventory", err.Error())
}
