package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidDebianGetLinuxDistro(t *testing.T) {
	distro, err := GetLinuxOS("testdata/get_linux_distro/valid_debian")
	if err != nil {
		t.Errorf("Should not have raised an error")
	}
	assert.Equal(t, distro, "debian")
}

func TestValidUbuntuGetLinuxDistro(t *testing.T) {
	distro, err := GetLinuxOS("testdata/get_linux_distro/valid_ubuntu")
	if err != nil {
		t.Errorf("Should not have raised an error")
	}
	assert.Equal(t, distro, "ubuntu")
}

func TestReadNonExistentFile(t *testing.T) {
	_, err := GetLinuxOS("path/that/will/never/exist")
	if err == nil {
		t.Errorf("Should have raised an error because the provided path does not exist")
	}
}

func TestInvalidVersionFile(t *testing.T) {
	_, err := GetLinuxOS("testdata/get_linux_distro/invalid")
	assert.Equal(t, err.Error(), "could not determine the os from the `testdata/get_linux_distro/invalid` file")
}
