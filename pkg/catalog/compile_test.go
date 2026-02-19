package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadNodeFromCode(t *testing.T) {
	resources, loadedRoles, tags, variables, err := CompileCatalog("testdata/valid_code", "dummy")
	if err != nil {
		t.Errorf("Could not load catalog from code : %s", err.Error())
	}
	assert.Equal(t, 2, len(resources))
	assert.Equal(t, 2, len(loadedRoles))
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, 0, len(variables))
}

func TestLoadNodeMissingGroup(t *testing.T) {
	_, _, _, _, err := CompileCatalog("testdata/missing_group", "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a group is missing")
	}
	assert.Equal(t, err.Error(), "The group web could not be found in the inventory")
}

func TestLoadNodeMissingRole(t *testing.T) {
	_, _, _, _, err := CompileCatalog("testdata/missing_role", "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.Equal(t, err.Error(), "The role test could not be found in the roles folder")
}

func TestLoadNodeMissingNode(t *testing.T) {
	_, _, _, _, err := CompileCatalog("testdata/missing_node", "dummy")
	if err == nil {
		t.Errorf("Should have returned an error because a role is missing")
	}
	assert.Equal(t, err.Error(), "The node dummy could not be found in the inventory")
}
