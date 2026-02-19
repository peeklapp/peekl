package facts

import (
	"fmt"
	"runtime"

	"github.com/peeklapp/peekl/pkg/facts/collectors"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
)

type Facter struct {
	distro string
}

func (f *Facter) collectPackage() ([]models.Package, error) {
	var packages []models.Package

	collectorMapping := map[string]string{
		"debian": "dpkg",
		"ubuntu": "dpkg",
	}

	packages, err := collectors.CollectPackagesBasedOnSource(collectorMapping[f.distro])
	if err != nil {
		return packages, fmt.Errorf("An error happening during collection of packages : %s", err.Error())
	}

	return packages, nil
}

func (f *Facter) collectHostnme() (string, error) {
	return collectors.GetHostname()
}

func (f *Facter) Collect() (*models.Facts, error) {
	var facts models.Facts

	// First we need to determine the OS
	distro, err := utils.GetLinuxOS("")
	if err != nil {
		return &facts, err
	}
	f.distro = distro
	facts.Os.Distro = distro
	facts.Os.Arch = runtime.GOARCH

	// Collect list of packages
	pkgs, err := f.collectPackage()
	if err != nil {
		return &facts, err
	}
	facts.Packages = pkgs

	// Collect hostname
	hostname, err := f.collectHostnme()
	if err != nil {
		return &facts, err
	}
	facts.Hostname = hostname

	return &facts, nil
}

func NewFacter() *Facter {
	var facter Facter
	return &facter
}
