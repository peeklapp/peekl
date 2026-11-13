package catalog

import (
	"testing"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOrderingRolesWithDependencies(t *testing.T) {
	rolesToOrder := []models.Role{
		{
			Name: "nginx",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"common", "apt"},
				},
			},
		},
		{
			Name: "apt",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"common"},
				},
			},
		},
		{
			Name: "common",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{},
				},
			},
		},
		{
			Name: "myapp",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"nginx"},
				},
			},
		},
	}

	expectedFinalOrder := []models.Role{
		{
			Name: "common",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{},
				},
			},
		},
		{
			Name: "apt",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"common"},
				},
			},
		},
		{
			Name: "nginx",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"common", "apt"},
				},
			},
		},
		{
			Name: "myapp",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"nginx"},
				},
			},
		},
	}

	orderedRoles, err := orderRoles(rolesToOrder)
	if err != nil {
		t.Errorf("No error should have happened: %s", err.Error())
	}
	assert.Equal(t, orderedRoles, expectedFinalOrder)
}

func TestDependencyCycle(t *testing.T) {
	rolesToOrder := []models.Role{
		{
			Name: "nginx",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"apt"},
				},
			},
		},
		{
			Name: "apt",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"nginx"},
				},
			},
		},
	}

	_, err := orderRoles(rolesToOrder)
	if err.Error() != "detected a cycle dependency with role : 'nginx'" {
		t.Errorf("Should have raised an error for a cycling dependency, but got error : %s", err.Error())
	}
}

func TestNonExistingRole(t *testing.T) {
	rolesToOrder := []models.Role{
		{
			Name: "nginx",
			Entrypoint: models.RoleEntrypoint{
				Resources:       []models.Resource{},
				LoadedResources: []models.LoadedResource{},
				Includes:        []models.RoleIncludeEntry{},
				Meta: &models.RoleMeta{
					Requires: []string{"apt"},
				},
			},
		},
	}

	_, err := orderRoles(rolesToOrder)
	if err.Error() != "detected a dependency to a role that either does not exist or is not imported for this node : 'apt'" {
		t.Errorf("Should have raised an error for a dependency on a role that doesn't exist, but got error : %s", err.Error())
	}
}
