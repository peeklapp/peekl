package facts

import (
	"github.com/peeklapp/peekl/pkg/facts/collectors"
	"github.com/peeklapp/peekl/pkg/models"
)

type Facter struct{}

func (f *Facter) Collect() (*models.Facts, error) {
	var facts models.Facts

	// Collect Distribution data
	distributionData, err := collectors.GetDistributionData()
	if err != nil {
		return &facts, err
	}
	facts.Distribution = distributionData

	// Collect list of packages
	pkgs, err := collectors.GetPackagesByDistro(facts.Distribution.Id)
	if err != nil {
		return &facts, err
	}
	facts.Packages = pkgs

	// Collect hostname
	hostname, err := collectors.GetHostname()
	if err != nil {
		return &facts, err
	}
	facts.Hostname = hostname

	// Collect network interfaces
	networkInterfaces, err := collectors.GetNetworkInterfaces()
	if err != nil {
		return &facts, err
	}
	facts.NetworkInterfaces = networkInterfaces

	// Collect DMI data
	dmiData, err := collectors.GetDmiData()
	if err != nil {
		return &facts, err
	}
	facts.Dmi = dmiData

	// Collect disks
	disks, err := collectors.GetDisks()
	if err != nil {
		return &facts, err
	}
	facts.Disks = disks

	return &facts, nil
}

func NewFacter() *Facter {
	return &Facter{}
}
