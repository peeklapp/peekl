package collectors

import (
	"fmt"
	"os"
	"strings"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/utils"
)

func getOsReleaseContent() (string, error) {
	if !utils.FileExist("/etc/os-release") {
		return "", fmt.Errorf("Error while trying to read file content for distribution facts data : '/etc/os-release' file does not exist")
	}

	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "", fmt.Errorf("Error while trying to read file content for distribution facts data : %s", err.Error())
	}

	return string(data), nil
}

func processOsReleaseContent(osReleaseContent string) models.DistributionData {
	distributionData := models.DistributionData{
		Name:    "",
		Version: "",
		Release: "",
		Id:      "",
	}

	splittedOutput := strings.SplitSeq(osReleaseContent, "\n")
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

	return distributionData
}

func GetDistributionData() (models.DistributionData, error) {
	var distributionData models.DistributionData
	osReleaseContent, err := getOsReleaseContent()
	if err != nil {
		return distributionData, err
	}
	distributionData = processOsReleaseContent(osReleaseContent)
	return distributionData, nil
}
