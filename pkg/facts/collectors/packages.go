package collectors

import (
	"fmt"

	"github.com/peeklapp/peekl/pkg/facts/collectors/dpkg"
	"github.com/peeklapp/peekl/pkg/models"
)

func GetPackages(distro string) ([]models.Package, error) {
	packageCollectorMapping := map[string]string{
		"Debian": "dpkg",
		"Ubuntu": "dpkg",
	}

	switch packageCollectorMapping[distro] {
	case "dpkg":
		pkgs, err := dpkg.GetInstalledPackagesList()
		if err != nil {
			return pkgs, fmt.Errorf("An error happened while getting list of installed packages using dpkg : %s", err.Error())
		}
		return pkgs, nil
	default:
		return nil, fmt.Errorf("Unknown package collection method for distro : %s", distro)
	}
}
