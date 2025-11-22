package catalog

import (
	"fmt"

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

type Catalog struct {
	resources []models.LoadedResource
	facts     *models.Facts
	context   *models.CatalogContext
}

// Run the catalog
func (c *Catalog) Process() error {
	var created int
	var deleted int
	var updated int
	var failed int
	var unchanged int

	var resContext models.ResourceContext
	resContext.Facts = c.facts

	logrus.Info(
		fmt.Sprintf(
			"Starting process of catalog with %d resources to process",
			len(c.resources),
		),
	)

	for _, res := range c.resources {
		result, err := res.Process(&resContext)
		if err != nil {
			return err
		}

		if result.Created {
			created = created + 1
		} else if result.Deleted {
			deleted = deleted + 1
		} else if result.Updated {
			updated = updated + 1
		} else if result.Failed {
			failed = failed + 1
		} else {
			unchanged = unchanged + 1
		}
	}

	logrus.Info(
		fmt.Sprintf(
			"Finished process of catalog with the following result : %d created / %d deleted / %d updated / %d failed / %d unchanged",
			created,
			deleted,
			updated,
			failed,
			unchanged,
		),
	)

	return nil
}

// Validate that the catalog is valid
func (c *Catalog) Validate() error {
	return nil
}

func (c *Catalog) loadResources(resources []models.Resource) error {
	for _, res := range resources {
		switch res.Type {
		case "builtin.user":
			userRes, err := user.NewUserResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, userRes)
		case "builtin.group":
			groupRes, err := group.NewGroupResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, groupRes)
		case "builtin.file":
			fileRes, err := file.NewFileResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, fileRes)
		case "builtin.directory":
			directoryRes, err := directory.NewDirectoryResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, directoryRes)
		case "builtin.pkg":
			pkgRes, err := pkg.NewPackageResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, pkgRes)
		case "builtin.template":
			// Create template resource
			templRes, err := template.NewTemplateResource(&res, c.context.GlobalTemplateDirectory)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, templRes)
		case "builtin.systemd_service":
			servRes, err := systemdservice.NewSystemdServiceResource(&res)
			if err != nil {
				return err
			}
			c.resources = append(c.resources, servRes)
		}
	}
	return nil
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

func NewCatalog(rawResources []models.Resource, facts *models.Facts, context models.CatalogContext) (*Catalog, error) {
	var catalog Catalog
	catalog.facts = facts
	catalog.context = &context

	// Sort catalog
	sortedResources, err := sortResources(rawResources)
	if err != nil {
		return &catalog, err
	}

	// Load resources
	err = catalog.loadResources(sortedResources)
	if err != nil {
		return &catalog, err
	}

	return &catalog, nil
}
