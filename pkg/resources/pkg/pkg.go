package pkg

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/redat00/peekl/pkg/models"
	"github.com/sirupsen/logrus"
)

type installerInterface interface {
	Install([]models.Package) error                   // Used to install a package
	Remove([]models.Package) error                    // Used to remove a package
	Upgrade([]models.Package) error                   // Used to upgrade/downgrade a package
	ListInstalledPackages() ([]models.Package, error) // Used to list installed packages
}

func pkgInListWithoutVersion(pkgToFindInList models.Package, pkgs []models.Package) bool {
	for _, pkg := range pkgs {
		if pkg.Name == pkgToFindInList.Name {
			return true
		}
	}
	return false
}

type PackageData struct {
	Names             []string `mapstructure:"names"`
	Provider          string   `mapstructure:"provider"`
	installer         installerInterface
	processedPackages []models.Package
}

type PackageResource struct {
	Title   string
	Type    string
	Present bool
	Data    PackageData
}

func (p *PackageResource) ProcessPackageList() {
	for _, pkg := range p.Data.Names {
		var name string
		var version string

		if strings.Contains(pkg, "=") {
			splitted := strings.Split(pkg, "=")
			name = splitted[0]
			version = splitted[1]
		} else {
			name = pkg
		}

		var pkg models.Package
		pkg.Name = name
		pkg.Version = version

		p.Data.processedPackages = append(p.Data.processedPackages, pkg)
	}
}

func (p *PackageResource) FilterPackagesStatus(installed []models.Package) []models.Package {
	var filteredPackages []models.Package
	for _, pkg := range p.Data.processedPackages {
		if p.Present {
			if !pkgInListWithoutVersion(pkg, installed) {
				filteredPackages = append(filteredPackages, pkg)
			}
		} else {
			if pkgInListWithoutVersion(pkg, installed) {
				filteredPackages = append(filteredPackages, pkg)
			}
		}
	}
	return filteredPackages
}

func (p *PackageResource) Process(context *models.ResourceContext) (models.ResourceResult, error) {
	var result models.ResourceResult
	p.ProcessPackageList()

	installedPackages, err := p.Data.installer.ListInstalledPackages()
	if err != nil {
		result.Failed = true
		return result, err
	}

	if p.Present {
		nonInstalledPackagesThatShouldBe := p.FilterPackagesStatus(installedPackages)
		var nonInstalledPackagesThatShouldBeNames []string
		for _, pkg := range nonInstalledPackagesThatShouldBe {
			nonInstalledPackagesThatShouldBeNames = append(nonInstalledPackagesThatShouldBeNames, pkg.Name)
		}
		if len(nonInstalledPackagesThatShouldBe) != 0 {
			logrus.Info(
				fmt.Sprintf(
					"Packages (%s) are not installed but should be",
					strings.Join(nonInstalledPackagesThatShouldBeNames, " "),
				),
			)
			err := p.Data.installer.Install(nonInstalledPackagesThatShouldBe)
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Created = true
			logrus.Info(
				fmt.Sprintf(
					"Packages (%s) have been installed",
					strings.Join(nonInstalledPackagesThatShouldBeNames, " "),
				),
			)
		}
		// Update the installed packages list
		installedPackages, err = p.Data.installer.ListInstalledPackages()

		// Find any packages without the good version
		var packagesWithWrongVersion []models.Package
		for _, pkg := range p.Data.processedPackages {
			// Skip any package for which version is not specified
			if pkg.Version != "" && !slices.Contains(installedPackages, pkg) {
				packagesWithWrongVersion = append(packagesWithWrongVersion, pkg)
			}
		}

		// If any package do not have the correct version
		if len(packagesWithWrongVersion) != 0 {
			// Prepare list of packages for output
			var packagesWithWrongVersionNames []string
			for _, pkg := range packagesWithWrongVersion {
				packagesWithWrongVersionNames = append(packagesWithWrongVersionNames, pkg.Name)
			}
			// Install packages with correct versions
			logrus.Info(
				fmt.Sprintf(
					"Packages (%s) does not have the correct version",
					strings.Join(packagesWithWrongVersionNames, " "),
				),
			)
			err := p.Data.installer.Upgrade(packagesWithWrongVersion)
			if err != nil {
				result.Failed = true
				return result, err
			}
			logrus.Info(
				fmt.Sprintf(
					"Packages (%s) versions have been updated",
					strings.Join(packagesWithWrongVersionNames, " "),
				),
			)
		}
	} else {
		installedPackagesThatShouldNot := p.FilterPackagesStatus(installedPackages)
		if len(installedPackagesThatShouldNot) != 0 {
			for _, pkg := range installedPackagesThatShouldNot {
				logrus.Info(
					fmt.Sprintf("Package (%s) is installed but should not", pkg.Name),
				)
			}
			err := p.Data.installer.Remove(installedPackagesThatShouldNot)
			if err != nil {
				result.Failed = true
				return result, err
			}
			result.Deleted = true
			for _, pkg := range installedPackagesThatShouldNot {
				logrus.Info(
					fmt.Sprintf("Package (%s) has been removed", pkg.Name),
				)
			}
		}
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
