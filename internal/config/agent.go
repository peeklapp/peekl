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

type AgentCertificateConfig struct {
	CaFilePath                string `mapstructure:"ca_file_path" yaml:"ca_file_path"`
	CsrFilePath               string `mapstructure:"csr_file_path" yaml:"csr_file_path"`
	CertificateFilePath       string `mapstructure:"certificate_file_path" yaml:"certificate_file_path"`
	CertificateKeyPath        string `mapstructure:"certificate_key_path" yaml:"certificate_key_path"`
	BootstrapPendingFilePath  string `mapstructure:"bootstrap_pending_file_path" yaml:"bootstrap_pending_file_path"`
	BootstrapCompleteFilePath string `mapstructure:"bootstrap_complete_file_path" yaml:"bootstrap_complete_file_path"`
}

type AgentDaemonConfig struct {
	LoopTime int `mapstructure:"loop_time" yaml:"loop_time"`
}

type AgentLoggingConfig struct {
	Format string `mapstructure:"format" yaml:"format"`
	Debug  bool   `mapstructure:"debug" yaml:"debug"`
}

type AgentConfig struct {
	Server       AgentServerConfig      `mapstructure:"server" yaml:"server"`
	Certificates AgentCertificateConfig `mapstructure:"certificates" yaml:"certificates"`
	Daemon       AgentDaemonConfig      `mapstructure:"daemon" yaml:"daemon"`
	Logging      AgentLoggingConfig     `mapstructure:"logging" yaml:"logging"`
	Environment  string                 `mapstructure:"environment" yaml:"environment"`
}

func NewAgentConfiguration(configFilePath string) (*AgentConfig, error) {
	var config AgentConfig

	// Define defaults values
	defaults := map[string]any{
		"server": map[string]any{
			"port": 9040,
			"host": "peekl",
		},
		"certificates": map[string]any{
			"ca_file_path":                 "/etc/peekl/ssl/ca/ca.pem",
			"csr_file_path":                "/etc/peekl/ssl/agent/agent.csr",
			"certificate_file_path":        "/etc/peekl/ssl/agent/agent.pem",
			"certificate_key_path":         "/etc/peekl/ssl/agent/agent.key",
			"bootstrap_pending_file_path":  "/etc/peekl/ssl/agent/.bootstrap_pending",
			"bootstrap_complete_file_path": "/etc/peekl/ssl/agent/.bootstrap_complete",
		},
		"daemon": map[string]any{
			"loop_time": 1800,
		},
		"logging": map[string]any{
			"format": "string",
			"debug":  false,
		},
		"environment": "production",
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

	return &config, nil
}
