package config

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

type ListenConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Host string `mapstructure:"host" yaml:"host"`
}

type CertificatesConfig struct {
	MainDirectory    string `mapstructure:"main_directory" yaml:"main_directory"`
	CaDirectory      string `mapstructure:"ca_directory" yaml:"ca_directory"`
	PendingDirectory string `mapstructure:"pending_directory" yaml:"pending_directory"`
	ServerDirectory  string `mapstructure:"server_directory" yaml:"server_directory"`
	SignedDirectory  string `mapstructure:"signed_directory" yaml:"signed_directory"`
	DatabasePath     string `mapstructure:"database_path" yaml:"database_path"`
}

type CodeConfig struct {
	Directory string `mapstructure:"directory" yaml:"directory"`
}

type LoggingConfig struct {
	Format string `mapstructure:"format" yaml:"format"`
	Debug  bool   `mapstructure:"debug" yaml:"debug"`
}

type ServerConfig struct {
	Listen       ListenConfig       `mapstructure:"listen" yaml:"listen"`
	Certificates CertificatesConfig `mapstructure:"certificates" yaml:"certificates"`
	Code         CodeConfig         `mapstructure:"code" yaml:"code"`
	Logging      LoggingConfig      `mapstructure:"logging" yaml:"logging"`
}

func NewServerConfiguration(configFilePath string) (*ServerConfig, error) {
	var config ServerConfig

	// Define defaults values
	defaults := map[string]any{
		"listen": map[string]any{
			"port": 9040,
			"host": "127.0.0.1",
		},
		"certificates": map[string]any{
			"main_directory":    "/etc/peekl/ssl",
			"ca_directory":      "/etc/peekl/ssl/ca",
			"pending_directory": "/etc/peekl/ssl/pending",
			"server_directory":  "/etc/peekl/ssl/server",
			"signed_directory":  "/etc/peekl/ssl/signed",
			"database_path":     "/etc/peekl/ssl/certs.db",
		},
		"code": map[string]any{
			"directory": "/etc/peekl/code",
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
