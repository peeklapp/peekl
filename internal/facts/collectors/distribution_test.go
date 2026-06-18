package collectors

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDistributionDataDebian(t *testing.T) {
	data, _ := os.ReadFile("testdata/os-release-debian")

	res := processOsReleaseContent(string(data))

	assert.Equal(t, "debian", res.Id)
	assert.Equal(t, "Debian GNU/Linux", res.Name)
	assert.Equal(t, "trixie", res.Release)
	assert.Equal(t, "13", res.Version)
}

func TestGetDistributionDataRocky(t *testing.T) {
	data, _ := os.ReadFile("testdata/os-release-rocky")

	res := processOsReleaseContent(string(data))

	assert.Equal(t, "rocky", res.Id)
	assert.Equal(t, "Rocky Linux", res.Name)
	assert.Equal(t, "10.1", res.Version)
	assert.Equal(t, "", res.Release)
}

func TestGetDistributionDataFedora(t *testing.T) {
	data, _ := os.ReadFile("testdata/os-release-fedora")

	res := processOsReleaseContent(string(data))

	assert.Equal(t, "fedora", res.Id)
	assert.Equal(t, "Fedora Linux", res.Name)
	assert.Equal(t, "44", res.Version)
	assert.Equal(t, "", res.Release)
}

func TestGetDistributionDataCentos(t *testing.T) {
	data, _ := os.ReadFile("testdata/os-release-centos")

	res := processOsReleaseContent(string(data))

	assert.Equal(t, "centos", res.Id)
	assert.Equal(t, "CentOS Stream", res.Name)
	assert.Equal(t, "10", res.Version)
	assert.Equal(t, "", res.Release)
}

func TestGetDistributionDataRedhat(t *testing.T) {
	data, _ := os.ReadFile("testdata/os-release-redhat")

	res := processOsReleaseContent(string(data))

	assert.Equal(t, "rhel", res.Id)
	assert.Equal(t, "Red Hat Enterprise Linux", res.Name)
	assert.Equal(t, "10.2", res.Version)
	assert.Equal(t, "", res.Release)
}
