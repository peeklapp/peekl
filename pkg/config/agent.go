package config

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type AgentServerConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Host string `mapstructure:"host" yaml:"host"`
}

type AgentCertificatesConfig struct {
	Directory string `mapstructure:"directory" yaml:"directory"`
}

type AgentLoggingConfig struct {
	Format string `mapstructure:"format" yaml:"format"`
	Debug  bool   `mapstructure:"debug" yaml:"debug"`
}

type AgentConfig struct {
	Listen       AgentServerConfig       `mapstructure:"server" yaml:"server"`
	Certificates AgentCertificatesConfig `mapstructure:"certificates" yaml:"certificates"`
	Logging      AgentLoggingConfig      `mapstructure:"logging" yaml:"logging"`
}

func NewAgentConfiguration(configFilePath string) (*AgentConfig, error) {
	var config AgentConfig

	// Define defaults values
	defaults := map[string]any{
		"server": map[string]any{
			"port": 9040,
			"host": "127.0.0.1",
		},
		"certificates": map[string]any{
			"directory": "/etc/peekl/ssl/agent",
		},
		"logging": map[string]any{
			"format": "string",
			"debug":  false,
		},
	}

	// Make default struct with default values
	err := mapstructure.Decode(defaults, &config)
	if err != nil {
		return &config, err
	}

	// Check if file exist
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		logrus.Error("No configuration file found at provided path, using default values")
		return &config, nil
	}

	// Read content of configuration file
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return &config, err
	}

	var rawYaml map[string]any
	err = yaml.Unmarshal(data, &rawYaml)
	if err != nil {
		return &config, err
	}

	// Override any defaults with the configuration file
	err = mapstructure.Decode(rawYaml, &config)

	return &config, nil
}
