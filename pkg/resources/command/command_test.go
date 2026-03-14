package command

import (
	"errors"
	"os"
	"testing"

	"github.com/peeklapp/peekl/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestCommand(t *testing.T) {
	t.Run("TestRunCommandOnce", func(t *testing.T) {
		present := true
		rawRes := models.Resource{
			Title:    "run_simple_command",
			Type:     "builtin.command",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"command": "echo test > /tmp/testing_file",
		}

		commandRes, err := NewCommandResource(&rawRes, data, nil)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		if _, err := os.Stat("/tmp/testing_file"); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("File /tmp/testing_file should not exist at that stage")
		}

		context := models.ResourceContext{}
		_, err = commandRes.Process(&context)
		if err != nil {
			t.Errorf("An error happened during the process of the resource : %s", err.Error())
		}

		if _, err := os.Stat("/tmp/testing_file"); errors.Is(err, os.ErrNotExist) {
			t.Errorf("File /tmp/testing_file should exist at that stage")
		}
	})
	t.Run("TestRunCommandTwice", func(t *testing.T) {
		present := true
		rawRes := models.Resource{
			Title:    "run_simple_command",
			Type:     "builtin.command",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"command": "echo test2 > /tmp/testing_file",
			"creates": "/tmp/testing_file",
		}

		commandRes, err := NewCommandResource(&rawRes, data, nil)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		fileData, _ := os.ReadFile("/tmp/testing_file")
		assert.Equal(t, "test\n", string(fileData))

		context := models.ResourceContext{}
		_, err = commandRes.Process(&context)
		if err != nil {
			t.Errorf("An error happened during the process of the resource : %s", err.Error())
		}

		fileData, _ = os.ReadFile("/tmp/testing_file")
		assert.Equal(t, "test\n", string(fileData))
	})
}
