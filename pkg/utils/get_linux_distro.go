package utils

import (
	"fmt"
	"os"
	"strings"
)

func GetLinuxOS(filePath string) (string, error) {
	if filePath == "" {
		filePath = "/proc/version"
	}

	// Get raw data from `/proc/version` file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert raw data to string data
	stringData := string(data)

	// Try to find OS in the data
	if strings.Contains(stringData, "Debian") {
		return "debian", nil
	} else if strings.Contains(stringData, "Ubuntu") {
		return "ubuntu", nil
	}

	// If we're here, we did not find anything
	return "", fmt.Errorf("could not determine the os from the `%s` file", filePath)
}
