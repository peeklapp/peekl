package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/internal/models"
)

type RepositoryConfig struct {
	Key string `yaml:"key"`
	Url string `yaml:"url"`
}

type CodeConfig struct {
	pathOfConfigurationFile string
	CodeFolder              string           `yaml:"code" mapstructure:"code"`
	StagingFolder           string           `yaml:"staging" mapstructure:"staging"`
	Repository              RepositoryConfig `yaml:"repository"`
}

func (c *CodeConfig) Validate() error {
	validationErrors := []models.ValidationError{}

	if c.Repository.Url == "" {
		validationErrors = append(validationErrors, models.ValidationError{
			FieldName:    "repository.url",
			ViolatedRule: "Field is required",
		})
	}

	if !strings.HasPrefix(c.Repository.Url, "http") && c.Repository.Key == "" {
		validationErrors = append(validationErrors, models.ValidationError{
			FieldName:    "repository.key",
			ViolatedRule: "Repository URL appears to be SSH. The key field is required in that case.",
		})
	}

	if len(validationErrors) > 0 {
		return models.ConfigurationValidationError{
			Path:             c.pathOfConfigurationFile,
			ValidationErrors: validationErrors,
		}
	}
	return nil
}

func NewCodeConfiguration(configFilePath string) (*CodeConfig, error) {
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("No configuration file found at provided path : %s", configFilePath)
	}

	var config CodeConfig

	config.pathOfConfigurationFile = configFilePath

	defaults := map[string]any{
		"code":    "/etc/peekl/code",
		"staging": "/var/lib/peekl/code",
	}

	err := mapstructure.Decode(defaults, &config)
	if err != nil {
		return &config, err
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return &config, err
	}

	var rawYaml map[string]any
	err = yaml.Unmarshal(data, &rawYaml)
	if err != nil {
		return &config, err
	}

	err = mapstructure.Decode(rawYaml, &config)
	if err != nil {
		return &config, err
	}

	err = config.Validate()
	if err != nil {
		return &config, err
	}

	return &config, nil
}
