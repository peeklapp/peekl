package code

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peeklapp/peekl/internal/utils"
)

func TestGenerateNodesArchives(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("testdata/", "nodes")
	if err != nil {
		t.Errorf("Returned an error during temporary directory creation : %s", err.Error())
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Generate archive
	err = GenerateNodesArchives("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during nodes archices creation : %s", err.Error())
	}

	// Assert that archive exist
	if !utils.FileExist(filepath.Join(tempDir, "testing.tar.zst"), nil) {
		t.Errorf("Archive does not exist after generation")
	}
}

func TestGenerateCodeArchive(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("testdata/", "code")
	if err != nil {
		t.Errorf("Returned an error during temporary directory creation : %s", err.Error())
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck

	// Generate archive
	err = GenerateCodeArchive("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during code archive creation : %s", err.Error())
	}

	// Assert that archive exist
	if !utils.FileExist(filepath.Join(tempDir, "code.tar.zst"), nil) {
		t.Errorf("Archive does not exist after generation")
	}
}
