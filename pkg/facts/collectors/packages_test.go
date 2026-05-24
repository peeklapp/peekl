package collectors

import (
	"os"
	"testing"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestLoadDpkgDump(t *testing.T) {
	data, _ := os.ReadFile("testdata/dpkg_dump.txt")

	res := processRawPackagesList(string(data))
	assert.Equal(t, 26, len(res))
	assert.Equal(t, models.Package{Name: "zstd", Version: "1.5.7+dfsg-1"}, res[24])
}

func TestLoadRpmDump(t *testing.T) {
	data, _ := os.ReadFile("testdata/rpm_dump.txt")

	res := processRawPackagesList(string(data))
	assert.Equal(t, 34, len(res))
	assert.Equal(t, models.Package{Name: "nginx", Version: "1.26.3.1.el10"}, res[33])

}
