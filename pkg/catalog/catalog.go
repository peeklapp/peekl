package catalog

import (
	"fmt"

	"github.com/redat00/peekl/pkg/models"
	"github.com/redat00/peekl/pkg/resources/directory"
	"github.com/redat00/peekl/pkg/resources/file"
	"github.com/redat00/peekl/pkg/resources/group"
	"github.com/redat00/peekl/pkg/resources/user"
	"github.com/sirupsen/logrus"
)

// A Catalog is a list of all the resources that are managed
// for a given node. It's the list of users you want to create,
// the files you want to create... etc.

type Catalog struct {
	resources []models.LoadedResource
	facts     *models.Facts
}

func (c *Catalog) Process() error {
	var created int
	var deleted int
	var updated int
	var failed int
	var unchanged int

	logrus.Info(
		fmt.Sprintf(
			"Starting process of catalog with %d resources to process",
			len(c.resources),
		),
	)

	for _, res := range c.resources {
		result, err := res.Process()
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

func (c *Catalog) Validate() error {
	return nil
}

func loadResources(resources []models.Resource) ([]models.LoadedResource, error) {
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

func NewCatalog(rawResources []models.Resource, facts *models.Facts) (*Catalog, error) {
	var catalog Catalog
	catalog.facts = facts

	// Sort catalog
	sortedResources, err := sortResources(rawResources)
	if err != nil {
		return &catalog, err
	}

	// Load resources
	loadedResources, err := loadResources(sortedResources)
	if err != nil {
		return &catalog, err
	}
	catalog.resources = loadedResources

	return &catalog, nil
}
