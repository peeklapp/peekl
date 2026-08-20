package catalog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadNodeFromCode(t *testing.T) {
	root, _ := os.OpenRoot("testdata/valid_code")
	resources, loadedRoles, tags, variables, err := CompileCatalog(root, "dummy")
	if err != nil {
		t.Errorf("Could not load catalog from code : %s", err.Error())
	}
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, 2, len(loadedRoles))
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, 0, len(variables))
}

func TestNoDuplicateRoles(t *testing.T) {
	root, _ := os.OpenRoot("testdata/duplicate_role")
	_, loadedRoles, _, _, err := CompileCatalog(root, "dummy")
	if err != nil {
		t.Errorf("Could not load catalog from code : %s", err.Error())
	}
	assert.Equal(t, 2, len(loadedRoles))
}

func TestLoadNodeMissingGroup(t *testing.T) {
	root, _ := os.OpenRoot("testdata/missing_group")
	_, _, _, _, err := CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a group is missing")
	}
	assert.Equal(t, err.Error(), "The group web could not be found in the inventory")
}

func TestLoadNodeMissingRole(t *testing.T) {
	root, _ := os.OpenRoot("testdata/missing_role")
	_, _, _, _, err := CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.Equal(t, err.Error(), "The role test could not be found in the roles folder")
}

func TestLoadNodeMissingNode(t *testing.T) {
	root, _ := os.OpenRoot("testdata/missing_node")
	_, _, _, _, err := CompileCatalog(root, "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.Equal(t, err.Error(), "The node dummy could not be found in the inventory")
}
