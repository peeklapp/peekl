package catalog

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/resources/directory"
	"github.com/redat00/peekl/pkg/resources/file"
	"github.com/redat00/peekl/pkg/resources/group"
	"github.com/redat00/peekl/pkg/resources/pkg"
	"github.com/redat00/peekl/pkg/resources/systemd_service"
	"github.com/redat00/peekl/pkg/resources/template"
	"github.com/redat00/peekl/pkg/resources/user"
	"github.com/sirupsen/logrus"
)

// A Catalog is a list of all the resources that are managed
// for a given node. It's the list of users you want to create,
// the files you want to create... etc.

func shouldNotSkipResource(res models.LoadedResource, resContext *models.ResourceContext) (bool, error) {
	// If no when is specified, then we always run
	if res.When() == "" {
		return true, nil
	}

	// Create env for when request with facts and vars
	env := map[string]any{"facts": resContext.Facts, "vars": resContext.Variables, "tags": resContext.Tags}

	// Compile the when request, and then execute it
	compiledWhen, err := expr.Compile(res.When(), expr.Env(env))
	if err != nil {
		return false, err
	}
	output, err := expr.Run(compiledWhen, env)
	if err != nil {
		return false, err
	}

	// Check if output is bool, if it is return it, else return error
	if b, ok := output.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("'when' condition for res '%s' did not return a boolean !", res.String())
}

func processResources(resources []models.LoadedResource, resContext *models.ResourceContext, catalogResult *CatalogResult) error {
	for _, res := range resources {
		logrus.Info(fmt.Sprintf("Processing resource %s", res.String()))

		// Process 'when' condition of resource
		skip, err := shouldNotSkipResource(res, resContext)
		if err != nil {
			return err
		}
		if !skip {
			logrus.Debug(fmt.Sprintf("Resource %s has been skipped", res.String()))
			catalogResult.Skipped = catalogResult.Skipped + 1
			continue
		}

		// Process resource
		result, err := res.Process(resContext)
		if err != nil {
			return err
		}

		// Process resource result
		if result.Created {
			catalogResult.Created = catalogResult.Created + 1
		} else if result.Deleted {
			catalogResult.Deleted = catalogResult.Deleted + 1
		} else if result.Updated {
			catalogResult.Updated = catalogResult.Updated + 1
		} else if result.Failed {
			catalogResult.Failed = catalogResult.Failed + 1
		} else {
			catalogResult.Unchanged = catalogResult.Unchanged + 1
		}
	}

	return nil
}

type Catalog struct {
	resources []models.LoadedResource
	variables map[string]any
	facts     *models.Facts
	tags      []string
	roles     []models.Role
}

type CatalogResult struct {
	Created   int
	Deleted   int
	Updated   int
	Failed    int
	Unchanged int
	Skipped   int
}

// Run the catalog
func (c *Catalog) Process() error {
	var catalogResult CatalogResult

	var resContext models.ResourceContext
	resContext.Facts = c.facts
	resContext.Variables = c.variables
	resContext.Tags = c.tags

	logrus.Info(
		fmt.Sprintf(
			"Starting process of catalog with %d global resources and %d roles.",
			len(c.resources),
			len(c.roles),
		),
	)

	// Handle global resources
	err := processResources(c.resources, &resContext, &catalogResult)
	if err != nil {
		return err
	}

	// Handle roles
	for _, role := range c.roles {
		// Handle main
		logrus.Info(fmt.Sprintf("Starting process of role %s", role.Name))
		err := processResources(role.LoadedResources, &resContext, &catalogResult)
		if err != nil {
			return err
		}

		// Handle included
		for _, included := range role.IncludedResources {
			err := processResources(included.LoadedResources, &resContext, &catalogResult)
			if err != nil {
				return err
			}
		}
	}

	logrus.Info(
		fmt.Sprintf(
			"Finished process of catalog with the following result : %d created / %d deleted / %d updated / %d failed / %d unchanged / %d skipped",
			catalogResult.Created,
			catalogResult.Deleted,
			catalogResult.Updated,
			catalogResult.Failed,
			catalogResult.Unchanged,
			catalogResult.Skipped,
		),
	)

	return nil
}

// Validate that the catalog is valid
func (c *Catalog) Validate() error {
	return nil
}

func (c *Catalog) loadRoles(roles []models.Role) error {
	for _, role := range roles {
		// Handle main resources
		sortedMainResources, err := sortResources(role.Resources)
		if err != nil {
			return err
		}
		loadedMainResources, err := c.loadResources(sortedMainResources, nil)
		if err != nil {
			return err
		}
		role.LoadedResources = loadedMainResources

		// Handle each included
		for key, include := range role.IncludedResources {
			sortedIncludedResources, err := sortResources(include.Resources)
			if err != nil {
				return err
			}
			loadedIncludedResources, err := c.loadResources(sortedIncludedResources, &models.RoleContext{Templates: role.Templates})
			include.LoadedResources = loadedIncludedResources
			role.IncludedResources[key] = include
		}
		c.roles = append(c.roles, role)
	}

	return nil
}

func (c *Catalog) loadResources(resources []models.Resource, roleContext *models.RoleContext) ([]models.LoadedResource, error) {
	var loadedResources []models.LoadedResource
	for _, res := range resources {
		switch res.Type {
		case "builtin.user":
			userRes, err := user.NewUserResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, userRes)
		case "builtin.group":
			groupRes, err := group.NewGroupResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, groupRes)
		case "builtin.file":
			fileRes, err := file.NewFileResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, fileRes)
		case "builtin.directory":
			directoryRes, err := directory.NewDirectoryResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, directoryRes)
		case "builtin.pkg":
			pkgRes, err := pkg.NewPackageResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, pkgRes)
		case "builtin.template":
			// Create template resource
			templRes, err := template.NewTemplateResource(&res, roleContext.Templates)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, templRes)
		case "builtin.systemd_service":
			servRes, err := systemdservice.NewSystemdServiceResource(&res)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, servRes)
		}
	}
	return loadedResources, nil
}

func createResourceRequireIndexMap(resources []models.Resource) map[string]int {
	indexMap := make(map[string]int, len(resources))
	for i, res := range resources {
		key := fmt.Sprintf("%s:%s", res.Title, res.Type)
		indexMap[key] = i
	}
	return indexMap
}

func sortResources(resources []models.Resource) ([]models.Resource, error) {
	var sortedResources []models.Resource

	indexMap := createResourceRequireIndexMap(resources)
	visited := make([]int, len(resources))

	var visit func(i int) error
	visit = func(i int) error {
		// See if the resource has already been visited
		switch visited[i] {
		case 2:
			return nil
		case 1:
			return fmt.Errorf("Dependency cycle")
		}

		visited[i] = 1
		reqKey := fmt.Sprintf("%s:%s", resources[i].Require.Title, resources[i].Require.Type)

		reqIdx, ok := indexMap[reqKey]
		if !ok && (resources[i].Require.Title != "" || resources[i].Require.Type != "") {
			return fmt.Errorf(
				"%s",
				fmt.Sprintf(
					"Resource %s/%s requires an unknown resource : %s/%s",
					resources[i].Type,
					resources[i].Title,
					resources[i].Require.Type,
					resources[i].Require.Title,
				),
			)
		}

		if resources[i].Require.Title != "" || resources[i].Require.Type != "" {
			if err := visit(reqIdx); err != nil {
				return err
			}
		}
		visited[i] = 2
		sortedResources = append(sortedResources, resources[i])
		return nil
	}

	for i := range resources {
		if visited[i] == 0 {
			if err := visit(i); err != nil {
				return sortedResources, err
			}
		}
	}

	return sortedResources, nil
}

func NewCatalog(rawCatalog models.RawCatalog) (*Catalog, error) {
	var catalog Catalog

	// Add facts and catalog context
	catalog.variables = rawCatalog.Variables
	catalog.facts = rawCatalog.Facts

	// Handle global resources
	sortedGlobalResources, err := sortResources(rawCatalog.GlobalResources)
	if err != nil {
		return &catalog, err
	}

	// Load global resources
	globalLoadedResources, err := catalog.loadResources(sortedGlobalResources, nil)
	catalog.resources = globalLoadedResources

	// Handle roles
	catalog.loadRoles(rawCatalog.Roles)

	return &catalog, nil
}
