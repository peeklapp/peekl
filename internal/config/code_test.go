package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadNonExistentConfiguration(t *testing.T) {
	_, err := NewCodeConfiguration("testdata/code/i_dont_exist.yml")
	if err == nil {
		t.Errorf("Should have returned an error because the file doesn't exist.")
	}
	assert.Equal(t, "No configuration file found at provided path : testdata/code/i_dont_exist.yml", err.Error())
}

func TestLoadConfiguration(t *testing.T) {
	config, err := NewCodeConfiguration("testdata/code/config.yml")
	if err != nil {
		t.Errorf("Should not have returned any error : %s", err.Error())
	}

	assert.Equal(t, "/etc/test/peekl/code", config.CodeFolder)
	assert.Equal(t, "git@github.com:user/repo.git", config.Repository.Url)
}

func TestLoadConfigurationWithoutDefaults(t *testing.T) {
	config, err := NewCodeConfiguration("testdata/code/config_without_defaults.yml")
	if err != nil {
		t.Errorf("Shoud not have returned any error : %s", err.Error())
	}
	assert.Equal(t, "/etc/peekl/code", config.CodeFolder)
}
