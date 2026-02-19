package group

import (
	"os/user"
	"testing"

	"github.com/peeklapp/peekl/pkg/models"
)

func TestCreateAndDeleteGroup(t *testing.T) {
	t.Run("CreateGroup", func(t *testing.T) {
		present := true
		rawRes := models.Resource{
			Title:    "create_group_peekl",
			Type:     "builtin.group",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"name": "peekl",
		}
		groupRes, err := NewGroupResource(&rawRes, data)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		// Check that group currently does not exist
		_, err = user.LookupGroup("peekl")
		if err == nil {
			t.Errorf("Group peekl should not exist at this stage")
		}

		context := models.ResourceContext{}
		groupRes.Process(&context)

		_, err = user.LookupGroup("peekl")
		if err != nil {
			t.Errorf("Group peekl should exist at this stage")
		}
	})

	t.Run("DeleteGroup", func(t *testing.T) {
		present := false
		rawRes := models.Resource{
			Title:    "delete_group_peekl",
			Type:     "builtin.group",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"name": "peekl",
		}
		groupRes, err := NewGroupResource(&rawRes, data)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		// Check that group currently does not exist
		_, err = user.LookupGroup("peekl")
		if err != nil {
			t.Errorf("Group peekl should exist at this stage")
		}

		context := models.ResourceContext{}
		groupRes.Process(&context)

		_, err = user.LookupGroup("peekl")
		if err == nil {
			t.Errorf("Group peekl should not exist at this stage")
		}
	})
}
