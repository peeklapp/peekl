package catalog

import (
	"testing"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestOrderingRolesWithDependencies(t *testing.T) {
	rolesToOrder := []models.Role{
		{
			Name:              "nginx",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"common", "apt"},
		},
		{
			Name:              "apt",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"common"},
		},
		{
			Name:              "common",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{},
		},
		{
			Name:              "myapp",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"nginx"},
		},
	}

	expectedFinalOrder := []models.Role{
		{
			Name:              "common",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{},
		},
		{
			Name:              "apt",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"common"},
		},
		{
			Name:              "nginx",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"common", "apt"},
		},
		{
			Name:              "myapp",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"nginx"},
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
			Name:              "nginx",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"apt"},
		},
		{
			Name:              "apt",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"nginx"},
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
			Name:              "nginx",
			Resources:         []models.Resource{},
			LoadedResources:   []models.LoadedResource{},
			IncludedResources: map[string]models.IncludedResources{},
			Variables:         map[string]any{},
			DependsOn:         []string{"apt"},
		},
	}

	_, err := orderRoles(rolesToOrder)
	if err.Error() != "detected a dependency to a role that either does not exist or is not imported for this node : 'apt'" {
		t.Errorf("Should have raised an error for a dependency on a role that doesn't exist, but got error : %s", err.Error())
	}
}
