package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigurationAgent(t *testing.T) {
	config, err := NewAgentConfiguration("testdata/agent/config.yml")
	if err != nil {
		t.Errorf("Shoud not have returned any error : %s", err.Error())
	}

	assert.Equal(t, "peekl.dev", config.Server.Host)
	assert.Equal(t, 27000, config.Server.Port)
	assert.Equal(t, "/var/lib/peekl/ssl/ca/ca.pem", config.Certificates.CaFilePath)
	assert.Equal(t, "/var/lib/peekl/ssl/agent/agent.csr", config.Certificates.CsrFilePath)
	assert.Equal(t, "/var/lib/peekl/ssl/agent/agent.pem", config.Certificates.CertificateFilePath)
	assert.Equal(t, "/var/lib/peekl/ssl/agent/agent.key", config.Certificates.CertificateKeyPath)
	assert.Equal(t, "/var/lib/peekl/ssl/agent/.bootstrap_pending", config.Certificates.BootstrapPendingFilePath)
	assert.Equal(t, "/var/lib/peekl/ssl/agent/.bootstrap_complete", config.Certificates.BootstrapCompleteFilePath)
	assert.Equal(t, 3600, config.Daemon.LoopTime)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, true, config.Logging.Debug)
	assert.Equal(t, "testing", config.Environment)
}

func TestLoadConfigurationWithoutDefaultsAgent(t *testing.T) {
	config, err := NewAgentConfiguration("testdata/agent/config_without_defaults.yml")
	if err != nil {
		t.Errorf("Shoud not have returned any error : %s", err.Error())
	}

	assert.Equal(t, "peekl", config.Server.Host)
	assert.Equal(t, 9040, config.Server.Port)
	assert.Equal(t, "/etc/peekl/ssl/ca/ca.pem", config.Certificates.CaFilePath)
	assert.Equal(t, "/etc/peekl/ssl/agent/agent.csr", config.Certificates.CsrFilePath)
	assert.Equal(t, "/etc/peekl/ssl/agent/agent.pem", config.Certificates.CertificateFilePath)
	assert.Equal(t, "/etc/peekl/ssl/agent/agent.key", config.Certificates.CertificateKeyPath)
	assert.Equal(t, "/etc/peekl/ssl/agent/.bootstrap_pending", config.Certificates.BootstrapPendingFilePath)
	assert.Equal(t, "/etc/peekl/ssl/agent/.bootstrap_complete", config.Certificates.BootstrapCompleteFilePath)
	assert.Equal(t, 1800, config.Daemon.LoopTime)
	assert.Equal(t, "string", config.Logging.Format)
	assert.Equal(t, false, config.Logging.Debug)
	assert.Equal(t, "production", config.Environment)
}
