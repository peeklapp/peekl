package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/peeklapp/peekl/internal/models"
	"github.com/sirupsen/logrus"
)

type ListenConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Host string `mapstructure:"host" yaml:"host"`
}

type CertificatesConfig struct {
	CaCertificateFilePath     string   `mapstructure:"ca_certificate_file_path" yaml:"ca_certificate_file_path"`
	CaCertificateKeyPath      string   `mapstructure:"ca_certificate_key_path" yaml:"ca_certificate_key_path"`
	ServerCertificateFilePath string   `mapstructure:"server_certificate_file_path" yaml:"server_certificate_file_path"`
	ServerCertificateKeyPath  string   `mapstructure:"server_certificate_key_path" yaml:"server_certificate_key_path"`
	BootstrapDoneFilePath     string   `mapstructure:"bootstrap_done_file_path" yaml:"bootstrap_done_file_path"`
	BootstrapDnsNames         []string `mapstructure:"bootstrap_dns_names" yaml:"bootstrap_dns_names"`
}

type ServerCodeConfig struct {
	Directory string `mapstructure:"directory" yaml:"directory"`
}

type LoggingConfig struct {
	Format  string `mapstructure:"format" yaml:"format"`
	Debug   bool   `mapstructure:"debug" yaml:"debug"`
	LogPath string `mapstructure:"log_path" yaml:"log_path"`
}

type DatabaseConfig struct {
	Type       string `mapstructure:"type" yaml:"type"`
	Path       string `mapstructure:"path" yaml:"path"`
	Host       string `mapstructure:"host" yaml:"host"`
	Port       int    `mapstructure:"port" yaml:"port"`
	Name       string `mapstructure:"name" yaml:"name"`
	Username   string `mapstructure:"username" yaml:"username"`
	Password   string `mapstructure:"password" yaml:"password"`
	DisableSsl bool   `mapstructure:"disable_ssl" yaml:"disable_ssl"`
}

func (d *DatabaseConfig) ToDSN() string {
	if d.Type == "sqlite" {
		return d.Path
	}

	switch d.Type {
	case "postgres":
		{
			sslState := ""
			if d.DisableSsl {
				sslState = "?sslmode=disable"
			}

			return fmt.Sprintf(
				"postgresql://%s:%s@%s:%d%s",
				d.Username,
				d.Password,
				d.Host,
				d.Port,
				sslState,
			)
		}
	}

	// We should never end in this case
	return ""
}

type ServerConfig struct {
	pathOfConfigurationFile string
	Listen                  ListenConfig       `mapstructure:"listen" yaml:"listen"`
	Certificates            CertificatesConfig `mapstructure:"certificates" yaml:"certificates"`
	Code                    ServerCodeConfig   `mapstructure:"code" yaml:"code"`
	Logging                 LoggingConfig      `mapstructure:"logging" yaml:"logging"`
	Database                DatabaseConfig     `mapstructure:"database" yaml:"database"`
}

func (c *ServerConfig) Validate() error {
	validationErrors := []models.ValidationError{}

	// Validate database configuration
	switch c.Database.Type {
	case "sqlite":
		{
			if c.Database.Path == "" {
				validationErrors = append(validationErrors, models.ValidationError{
					FieldName:    "database.path",
					ViolatedRule: "Required when type is `sqlite`",
				})
			}
		}
	case "postgres":
		{
			if c.Database.Host == "" {
				validationErrors = append(validationErrors, models.ValidationError{
					FieldName:    "database.host",
					ViolatedRule: "Required when type is `postgres`",
				})

			}
			if c.Database.Port == 0 {
				validationErrors = append(validationErrors, models.ValidationError{
					FieldName:    "database.port",
					ViolatedRule: "Required when type is `postgres`",
				})

			}
			if c.Database.Username == "" {
				validationErrors = append(validationErrors, models.ValidationError{
					FieldName:    "database.username",
					ViolatedRule: "Required when type is `postgres`",
				})

			}
			if c.Database.Password == "" {
				validationErrors = append(validationErrors, models.ValidationError{
					FieldName:    "database.password",
					ViolatedRule: "Required when type is `postgres`",
				})
			}
		}
	default:
		{
			validationErrors = append(validationErrors, models.ValidationError{
				FieldName: "database.type",
				ViolatedRule: fmt.Sprintf(
					"%s is an unknown database type, accepted values are : sqlite, postgres",
					c.Database.Type,
				),
			},
			)
		}
	}

	if len(validationErrors) > 0 {
		return models.ConfigurationValidationError{
			Path:             c.pathOfConfigurationFile,
			ValidationErrors: validationErrors,
		}
	}
	return nil
}

func NewServerConfiguration(configFilePath string) (*ServerConfig, error) {
	var config ServerConfig

	config.pathOfConfigurationFile = configFilePath

	// Define defaults values
	defaults := map[string]any{
		"listen": map[string]any{
			"port": 9040,
			"host": "127.0.0.1",
		},
		"certificates": map[string]any{
			"ca_certificate_file_path":     "/etc/peekl/ssl/ca/ca.pem",
			"ca_certificate_key_path":      "/etc/peekl/ssl/ca/ca.key",
			"server_certificate_file_path": "/etc/peekl/ssl/server/server.pem",
			"server_certificate_key_path":  "/etc/peekl/ssl/server/server.key",
			"bootstrap_done_file_path":     "/etc/peekl/ssl/server/.bootstrap_done",
			"bootstrap_dns_names":          []string{},
		},
		"code": map[string]any{
			"directory": "/etc/peekl/code",
		},
		"logging": map[string]any{
			"format":   "string",
			"debug":    false,
			"log_path": "/var/log/peekl/peekl.log",
		},
		"database": map[string]any{
			"type": "sqlite",
			"path": "/var/lib/peekl/data.db",
			"name": "peekl",
		},
	}

	// Make default struct with default values
	err := mapstructure.Decode(defaults, &config)
	if err != nil {
		return &config, err
	}

	// Check if file exist
	if _, err := os.Stat(configFilePath); errors.Is(err, os.ErrNotExist) {
		logrus.Warn("No configuration file found at provided path, using default values")
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

	// Validate the configuration
	if err = config.Validate(); err != nil {
		return &config, err
	}

	return &config, nil
}
