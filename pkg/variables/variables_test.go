package variables

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportRoleVariables(t *testing.T) {
	t.Run("ImportValidRole", func(t *testing.T) {
		variables, err := LoadRoleVariables("testdata/", "test_role")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{"test": "test", "dict": map[string]any{"bonjour": "hello", "hola": "bonjour"}}, variables)
	})

	t.Run("ImportEmptyRole", func(t *testing.T) {
		variables, err := LoadRoleVariables("testdata/", "test_empty_role")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{}, variables)
	})
}

func TestImportGroupVariables(t *testing.T) {
	t.Run("ImportValidGroup", func(t *testing.T) {
		variables, err := LoadGroupVariables("testdata/", "test_group")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{"test": "test", "dict": map[string]any{"bonjour": "hello", "hola": "bonjour"}}, variables)
	})

	t.Run("ImportEmptyGroup", func(t *testing.T) {
		variables, err := LoadGroupVariables("testdata/", "test_empty_group")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{}, variables)
	})
}

func TestImportNodeVariables(t *testing.T) {
	t.Run("ImportValidNode", func(t *testing.T) {
		variables, err := LoadNodeVariables("testdata/", "test_node")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{"test": "test", "dict": map[string]any{"bonjour": "hello", "hola": "bonjour"}}, variables)
	})

	t.Run("ImportEmptyNode", func(t *testing.T) {
		variables, err := LoadNodeVariables("testdata/", "test_empty_node")
		if err != nil {
			t.Errorf("Should not raise any error : %s", err.Error())
		}
		assert.Equal(t, map[string]any{}, variables)
	})
}
