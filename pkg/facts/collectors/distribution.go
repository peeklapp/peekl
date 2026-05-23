package collectors

import (
	"fmt"
	"os"
	"strings"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
)

func GetDistributionData() (models.DistributionData, error) {
	var distributionData models.DistributionData

	if !utils.FileExist("/etc/os-release") {
		return distributionData, nil
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return distributionData, fmt.Errorf("Error while trying to read file content for distribution facts data : %s", err.Error())
	}

	splittedOutput := strings.SplitSeq(string(data), "\n")
	for line := range splittedOutput {
		if line != "" && strings.Contains(line, "=") {
			splittedLine := strings.Split(line, "=")
			trimmedValue := strings.Trim(splittedLine[1], "\"")
			switch splittedLine[0] {
			case "NAME":
				distributionData.Name = trimmedValue
			case "VERSION_ID":
				distributionData.Version = trimmedValue
			case "VERSION_CODENAME":
				distributionData.Release = trimmedValue
			case "ID":
				distributionData.Id = trimmedValue
			}
		}
	}

	return distributionData, nil
}
