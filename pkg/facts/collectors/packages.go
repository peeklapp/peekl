package collectors

import (
	"fmt"
	"strings"

	"github.com/peeklapp/peekl/pkg/facts/collectors/dpkg"
	"github.com/peeklapp/peekl/pkg/facts/collectors/rpm"
	"github.com/peeklapp/peekl/pkg/models"
)

func processRawPackagesList(rawPackageList string) []models.Package {
	var pkgs []models.Package
	splittedOutput := strings.SplitSeq(rawPackageList, "\n")
	for line := range splittedOutput {
		if line != "" {
			var pkg models.Package
			splittedLine := strings.Split(line, ";")
			pkg.Name = splittedLine[0]
			pkg.Version = splittedLine[1]
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs
}

func GetPackagesByCollector(collector string) ([]models.Package, error) {
	var rawPackageList string
	var err error

	switch collector {
	case "rpm":
		rawPackageList, err = dpkg.GetInstalledPackagesList()
	case "dpkg":
		rawPackageList, err = rpm.GetInstalledPackagesList()
	default:
		return nil, fmt.Errorf("Unknown package collection method : %s", collector)
	}

	if err != nil {
		return nil, err
	}

	return processRawPackagesList(rawPackageList), nil
}

func GetPackagesByDistro(distro string) ([]models.Package, error) {
	packageCollectorMapping := map[string]string{
		"debian": "dpkg",
		"ubuntu": "dpkg",
		"rocky":  "rpm",
	}

	var pkgs []models.Package
	var err error

	switch packageCollectorMapping[distro] {
	case "dpkg":
		pkgs, err = GetPackagesByCollector("dpkg")
	case "rpm":
		pkgs, err = GetPackagesByCollector("rpm")
	default:
		return nil, fmt.Errorf("Could not find any collector for provided distro : %s", err)
	}

	if err != nil {
		return nil, err
	}

	return pkgs, nil
}
