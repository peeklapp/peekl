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
