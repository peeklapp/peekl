package pkg

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/sirupsen/logrus"
)

type installerInterface interface {
	Install(string, string) error
	Remove(string) error
	IsPackageInstalled(string) bool
}

type PackageData struct {
	Name      string `mapstructure:"name"`
	Version   string `mapstructure:"version"`
	Provider  string `mapstructure:"provider"`
	installer installerInterface
}

type PackageResource struct {
	Title   string
	Type    string
	Present bool
	Data    PackageData
}

func (p *PackageResource) Process(context *models.Context) (models.ResourceResult, error) {
	var result models.ResourceResult

	if !p.Data.installer.IsPackageInstalled(p.Data.Name) && p.Present {
		logrus.Info(
			fmt.Sprintf("Package (%s) not installed, but should", p.Data.Name),
		)
		err := p.Data.installer.Install(p.Data.Name, p.Data.Version)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Package (%s) installed", p.Data.Name),
		)
		result.Created = true
	} else if p.Data.installer.IsPackageInstalled(p.Data.Name) && !p.Present {
		logrus.Info(
			fmt.Sprintf("Package (%s) is installed, but should not", p.Data.Name),
		)
		err := p.Data.installer.Remove(p.Data.Name)
		if err != nil {
			result.Failed = true
			return result, err
		}
		logrus.Info(
			fmt.Sprintf("Package (%s) removed", p.Data.Name),
		)
		result.Deleted = true
	}

	return result, nil
}

func (p *PackageResource) String() string {
	return fmt.Sprintf("%s/%s", p.Type, p.Title)
}

func NewPackageResource(resource *models.Resource) (*PackageResource, error) {
	var packageResource PackageResource

	defaults := map[string]any{
		"provider": "apt",
	}

	var packageData PackageData

	err := mapstructure.Decode(defaults, &packageData)
	if err != nil {
		return &packageResource, err
	}

	err = mapstructure.Decode(resource.Data, &packageData)
	if err != nil {
		return &packageResource, err
	}

	var installer installerInterface
	switch packageData.Provider {
	case "apt":
		installer = AptInstaller{}
	}
	packageData.installer = installer

	packageResource.Title = resource.Title
	packageResource.Type = resource.Type
	packageResource.Present = resource.Present
	packageResource.Data = packageData

	return &packageResource, nil
}
