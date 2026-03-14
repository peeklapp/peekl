package user

import (
	"os/user"
	"testing"

	"github.com/peeklapp/peekl/pkg/models"
)

func TestCreateAndDeleteUser(t *testing.T) {
	t.Run("CreateUser", func(t *testing.T) {
		present := true
		rawRes := models.Resource{
			Title:    "create_user_max",
			Type:     "builtin.user",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"username": "max",
		}
		userRes, err := NewUserResource(&rawRes, data, nil)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		_, err = user.Lookup("max")
		if err == nil {
			t.Errorf("User max should not exist at that stage")
		}

		context := models.ResourceContext{}
		_, err = userRes.Process(&context)
		if err != nil {
			t.Errorf("An error happened during the creation of the user: %s", err.Error())
		}

		_, err = user.Lookup("max")
		if err != nil {
			t.Errorf("User max should exist at that stage")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		present := false
		rawRes := models.Resource{
			Title:    "delete_user_max",
			Type:     "builtin.user",
			Present:  &present,
			When:     "",
			Register: "",
		}
		data := map[string]any{
			"username": "max",
		}
		userRes, err := NewUserResource(&rawRes, data, nil)
		if err != nil {
			t.Errorf("No error should happen at that stage")
		}

		_, err = user.Lookup("max")
		if err != nil {
			t.Errorf("User max should exist at that stage")
		}

		context := models.ResourceContext{}
		_, err = userRes.Process(&context)
		if err != nil {
			t.Errorf("An error happened during the deletion of the user: %s", err.Error())
		}

		_, err = user.Lookup("max")
		if err == nil {
			t.Errorf("User max should not exist at that stage")
		}
	})
}
