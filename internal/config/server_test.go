package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigurationServer(t *testing.T) {
	config, err := NewServerConfiguration("testdata/server/config.yml")
	if err != nil {
		t.Errorf("Shoud not have returned any error : %s", err.Error())
	}

	assert.Equal(t, "192.168.1.1", config.Listen.Host)
	assert.Equal(t, 27000, config.Listen.Port)
	assert.Equal(t, "/var/lib/peekl/ca.pem", config.Certificates.CaCertificateFilePath)
	assert.Equal(t, "/var/lib/peekl/ca.key", config.Certificates.CaCertificateKeyPath)
	assert.Equal(t, "/var/lib/peekl/server.pem", config.Certificates.ServerCertificateFilePath)
	assert.Equal(t, "/var/lib/peekl/server.key", config.Certificates.ServerCertificateKeyPath)
	assert.Equal(t, "/var/lib/peekl/done_bootstrap", config.Certificates.BootstrapDoneFilePath)
	assert.Equal(t, []string{"peekl.dev"}, config.Certificates.BootstrapDnsNames)
	assert.Equal(t, "/var/lib/code", config.Code.Directory)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, true, config.Logging.Debug)
	assert.Equal(t, "/var/log/main/peekl.log", config.Logging.LogPath)
	assert.Equal(t, "postgres", config.Database.Type)
	assert.Equal(t, "localhost", config.Database.Host)
	assert.Equal(t, 5432, config.Database.Port)
	assert.Equal(t, "dummy_user", config.Database.Username)
	assert.Equal(t, "dummy_password", config.Database.Password)
}

func TestLoadConfigurationWithoutDefaultsServer(t *testing.T) {
	config, err := NewServerConfiguration("testdata/server/config_without_defaults.yml")
	if err != nil {
		t.Errorf("Shoud not have returned any error : %s", err.Error())
	}

	assert.Equal(t, "127.0.0.1", config.Listen.Host)
	assert.Equal(t, 9040, config.Listen.Port)
	assert.Equal(t, "/etc/peekl/ssl/ca/ca.pem", config.Certificates.CaCertificateFilePath)
	assert.Equal(t, "/etc/peekl/ssl/ca/ca.key", config.Certificates.CaCertificateKeyPath)
	assert.Equal(t, "/etc/peekl/ssl/server/server.pem", config.Certificates.ServerCertificateFilePath)
	assert.Equal(t, "/etc/peekl/ssl/server/server.key", config.Certificates.ServerCertificateKeyPath)
	assert.Equal(t, "/etc/peekl/ssl/server/.bootstrap_done", config.Certificates.BootstrapDoneFilePath)
	assert.Equal(t, []string{}, config.Certificates.BootstrapDnsNames)
	assert.Equal(t, "/etc/peekl/code", config.Code.Directory)
	assert.Equal(t, "string", config.Logging.Format)
	assert.Equal(t, false, config.Logging.Debug)
	assert.Equal(t, "/var/log/peekl/peekl.log", config.Logging.LogPath)
	assert.Equal(t, "sqlite", config.Database.Type)
	assert.Equal(t, "/var/lib/peekl/data.db", config.Database.Path)
	assert.Equal(t, "peekl", config.Database.Name)
}
