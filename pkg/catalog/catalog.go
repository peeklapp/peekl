package catalog

import (
	"errors"
	"fmt"
	"maps"

	"github.com/expr-lang/expr"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/resources/command"
	"github.com/peeklapp/peekl/pkg/resources/debug"
	"github.com/peeklapp/peekl/pkg/resources/directory"
	"github.com/peeklapp/peekl/pkg/resources/file"
	"github.com/peeklapp/peekl/pkg/resources/group"
	"github.com/peeklapp/peekl/pkg/resources/pkg"
	systemdDaemon "github.com/peeklapp/peekl/pkg/resources/systemd_daemon"
	"github.com/peeklapp/peekl/pkg/resources/systemd_service"
	"github.com/peeklapp/peekl/pkg/resources/template"
	"github.com/peeklapp/peekl/pkg/resources/user"
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
	env := map[string]any{}
	maps.Copy(env, resContext.Variables)
	env["facts"] = resContext.Facts
	env["tags"] = resContext.Tags

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
		logrus.Debug(fmt.Sprintf("[%s] Processing resource", res.String()))

		// Process 'when' condition of resource
		skip, err := shouldNotSkipResource(res, resContext)
		if err != nil {
			return err
		}
		if !skip {
			logrus.Debug(fmt.Sprintf("[%s] Resource has been skipped", res.String()))
			catalogResult.Skipped = catalogResult.Skipped + 1
			continue
		}

		// Output log for indication of resource finished being processed
		finishedOutputLog := fmt.Sprintf("[%s] Finished processing, result: ", res.String())

		// Process resource
		result, err := res.Process(resContext)
		if err != nil {
			logrus.Error(finishedOutputLog + "Failed")
			logrus.Error(fmt.Sprintf("[%s] %s", res.String(), err.Error()))
			return err
		}

		// Process 'register' field of resource
		registerField := res.Register()

		// Process resource result
		if result.Created {
			catalogResult.Created = catalogResult.Created + 1
			finishedOutputLog = finishedOutputLog + "Created"
			if registerField != "" {
				resContext.Variables[registerField] = "created"
			}
			logrus.Info(finishedOutputLog)
		} else if result.Deleted {
			catalogResult.Deleted = catalogResult.Deleted + 1
			finishedOutputLog = finishedOutputLog + "Deleted"
			if registerField != "" {
				resContext.Variables[registerField] = "deleted"
			}
			logrus.Info(finishedOutputLog)
		} else if result.Updated {
			catalogResult.Updated = catalogResult.Updated + 1
			finishedOutputLog = finishedOutputLog + "Updated"
			if registerField != "" {
				resContext.Variables[registerField] = "updated"
			}
			logrus.Info(finishedOutputLog)
		} else if result.Failed {
			catalogResult.Failed = catalogResult.Failed + 1
			finishedOutputLog = finishedOutputLog + "Failed"
			if registerField != "" {
				resContext.Variables[registerField] = "failed"
			}
			logrus.Info(finishedOutputLog)
		} else {
			catalogResult.Unchanged = catalogResult.Unchanged + 1
			finishedOutputLog = finishedOutputLog + "Unchanged"
			if registerField != "" {
				resContext.Variables[registerField] = "unchanged"
			}
			logrus.Debug(finishedOutputLog)
		}
	}

	return nil
}

type Catalog struct {
	resources   []models.LoadedResource
	variables   map[string]any
	facts       *models.Facts
	tags        []string
	roles       []models.Role
	environment string
	apiClient   models.ApiClient
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
	resContext.Environment = c.environment
	resContext.ApiClient = c.apiClient

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
		return nil
	}

	// Handle roles
	for _, role := range c.roles {
		logrus.Debug(fmt.Sprintf("[%s] Process the role", role.Name))

		// Handle main of role
		err := processResources(role.LoadedResources, &resContext, &catalogResult)
		if err != nil {
			break
		}

		// Handle included of role
		for _, included := range role.IncludedResources {
			err := processResources(included.LoadedResources, &resContext, &catalogResult)
			if err != nil {
				break
			}
		}

		logrus.Debug(fmt.Sprintf("[%s] Finished processing role", role.Name))
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
func (c *Catalog) Validate() (bool, error) {
	valid := true

	// Validate global catalog resources
	for _, res := range c.resources {
		err := res.Validate()
		if err != nil {
			if errors.As(err, &models.ResourceValidationError{}) {
				valid = false
				logrus.Error(err)
			} else {
				return false, err
			}
		}
	}

	// Validates roles
	for _, role := range c.roles {
		for _, res := range role.LoadedResources {
			err := res.Validate()
			if err != nil {
				if errors.As(err, &models.ResourceValidationError{}) {
					valid = false
					logrus.Error(err)
				} else {
					return false, err
				}
			}
		}
	}

	return valid, nil
}

func (c *Catalog) loadRoles(roles []models.Role) error {
	for _, role := range roles {
		// Handle main resources
		loadedMainResources, err := c.loadResources(role.Resources, &models.RoleContext{RoleName: role.Name})
		if err != nil {
			return fmt.Errorf("Failed loading resource in role %s : %s", role.Name, err.Error())
		}
		role.LoadedResources = loadedMainResources

		// Handle each included
		for key, include := range role.IncludedResources {
			loadedIncludedResources, err := c.loadResources(include.Resources, &models.RoleContext{RoleName: role.Name})
			if err != nil {
				return fmt.Errorf("Failed loading resource in role %s : %s", role.Name, err.Error())
			}
			include.LoadedResources = loadedIncludedResources
			role.IncludedResources[key] = include
		}
		c.roles = append(c.roles, role)
	}

	return nil
}

func (c *Catalog) loadSingleResource(resource models.Resource, dataField map[string]any, roleContext *models.RoleContext) (models.LoadedResource, error) {
	if resource.Present == nil {
		defaultPresentValue := true
		resource.Present = &defaultPresentValue
	}

	switch resource.Type {
	case "builtin.user":
		return user.NewUserResource(&resource, dataField, roleContext)
	case "builtin.group":
		return group.NewGroupResource(&resource, dataField, roleContext)
	case "builtin.file":
		return file.NewFileResource(&resource, dataField, roleContext)
	case "builtin.directory":
		return directory.NewDirectoryResource(&resource, dataField, roleContext)
	case "builtin.pkg":
		return pkg.NewPackageResource(&resource, dataField, roleContext)
	case "builtin.template":
		return template.NewTemplateResource(&resource, dataField, roleContext)
	case "builtin.systemd_service":
		return systemdService.NewSystemdServiceResource(&resource, dataField, roleContext)
	case "builtin.systemd_daemon":
		return systemdDaemon.NewSystemdDaemonResource(&resource, dataField, roleContext)
	case "builtin.debug":
		return debug.NewDebugResource(&resource, dataField, roleContext)
	case "builtin.command":
		return command.NewCommandResource(&resource, dataField, roleContext)
	}
	return nil, fmt.Errorf("Unknown resource type : %s", resource.Type)
}

func (c *Catalog) loadResources(resources []models.Resource, roleContext *models.RoleContext) ([]models.LoadedResource, error) {
	var loadedResources []models.LoadedResource
	for _, res := range resources {
		if len(res.With) > 0 {
			for dataFieldId := range res.With {
				loadedRes, err := c.loadSingleResource(res, res.With[dataFieldId], roleContext)
				if err != nil {
					return loadedResources, err
				}
				loadedResources = append(loadedResources, loadedRes)
			}
		} else {
			loadedRes, err := c.loadSingleResource(res, res.Data, roleContext)
			if err != nil {
				return loadedResources, err
			}
			loadedResources = append(loadedResources, loadedRes)
		}
	}
	return loadedResources, nil
}

func NewCatalog(rawCatalog models.RawCatalog) (*Catalog, error) {
	var catalog Catalog

	// Add facts and catalog context
	catalog.variables = rawCatalog.Variables
	catalog.facts = rawCatalog.Facts
	catalog.environment = rawCatalog.Environment
	catalog.apiClient = rawCatalog.ApiClient

	// Load global resources
	globalLoadedResources, err := catalog.loadResources(rawCatalog.GlobalResources, nil)
	if err != nil {
		return &catalog, err
	}
	catalog.resources = globalLoadedResources

	// Handle roles
	err = catalog.loadRoles(rawCatalog.Roles)
	if err != nil {
		return &catalog, err
	}

	return &catalog, nil
}
