package command

import (
	"os"
	"testing"

	"github.com/peeklapp/peekl/internal/models"
	"github.com/peeklapp/peekl/internal/utils"
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
		parameters := map[string]any{
			"command": "echo test > /tmp/testing_file",
		}

		commandRes, err := NewCommandResource(&rawRes, parameters, nil)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		if utils.FileExist("/tmp/testing_file", nil) {
			t.Errorf("File /tmp/testing_file should not exist at that stage")
		}

		context := models.ResourceContext{}
		_, err = commandRes.Process(&context)
		if err != nil {
			t.Errorf("An error happened during the process of the resource : %s", err.Error())
		}

		if !utils.FileExist("/tmp/testing_file", nil) {
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
		parameters := map[string]any{
			"command": "echo test2 > /tmp/testing_file",
			"creates": "/tmp/testing_file",
		}

		commandRes, err := NewCommandResource(&rawRes, parameters, nil)
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
