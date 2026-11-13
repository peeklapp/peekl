package code

import (
	"path/filepath"
	"testing"

	"github.com/peeklapp/peekl/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateNodesArchives(t *testing.T) {
	tempDir := t.TempDir()
	err := GenerateNodesArchives("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during nodes archices creation : %s", err.Error())
	}

	// Assert that archive exist
	if !utils.FileExist(filepath.Join(tempDir, "supernode.tar.zst"), nil) {
		t.Errorf("Archive does not exist after generation")
	}
}

func TestGenerateCodeArchive(t *testing.T) {
	tempDir := t.TempDir()
	err := GenerateCodeArchive("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during code archive creation : %s", err.Error())
	}

	// Assert that archive exist
	if !utils.FileExist(filepath.Join(tempDir, "code.tar.zst"), nil) {
		t.Errorf("Archive does not exist after generation")
	}
}

func TestNodeArchiveContent(t *testing.T) {
	tempDir := t.TempDir()
	if err := GenerateNodesArchives("testdata/uncompressed", tempDir); err != nil {
		t.Fatalf("Returned an error during nodes archives creation : %s", err.Error())
	}

	extractDir := t.TempDir()
	if err := DecompressArchive(filepath.Join(tempDir, "supernode.tar.zst"), extractDir); err != nil {
		t.Fatalf("Could not decompress node archive : %s", err.Error())
	}

	for _, expected := range []string{
		"inventory/nodes/supernode.hcl",
		"inventory/groups/common.hcl",
		"variables/nodes/supernode/main.hcl",
		"variables/groups/common/main.hcl",
	} {
		if !utils.FileExist(filepath.Join(extractDir, expected), nil) {
			t.Errorf("Node archive should contain '%s'", expected)
		}
	}
}

func TestEntityDirectoryRegex(t *testing.T) {
	re := entityDirectoryRegex("variables/nodes", "web")
	assert.True(t, re.MatchString("variables/nodes/web"))
	assert.True(t, re.MatchString("variables/nodes/web/main.hcl"))
	assert.False(t, re.MatchString("variables/nodes/web2/main.hcl"))
	assert.False(t, entityDirectoryRegex("variables/nodes", "w.b").MatchString("variables/nodes/web/main.hcl"))
}
