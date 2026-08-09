package code

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateNodesArchives(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("testdata/", "nodes")
	if err != nil {
		t.Errorf("Returned an error during temporary directory creation : %s", err.Error())
	}
	defer os.RemoveAll(tempDir)

	// Generate archive
	err = GenerateNodesArchives("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during nodes archices creation : %s", err.Error())
	}

	// Assert that archive exist
	if _, err := os.Stat(filepath.Join(tempDir, "testing.tar.zst")); errors.Is(err, os.ErrNotExist) {
		t.Errorf("Archive does not exist after generation")
	}
}

func TestGenerateCodeArchive(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("testdata/", "code")
	if err != nil {
		t.Errorf("Returned an error during temporary directory creation : %s", err.Error())
	}
	defer os.RemoveAll(tempDir)

	// Generate archive
	err = GenerateCodeArchive("testdata/uncompressed", tempDir)
	if err != nil {
		t.Errorf("Returned an error during code archive creation : %s", err.Error())
	}

	// Assert that archive exist
	if _, err := os.Stat(filepath.Join(tempDir, "code.tar.zst")); errors.Is(err, os.ErrNotExist) {
		t.Errorf("Archive does not exist after generation")
	}
}
