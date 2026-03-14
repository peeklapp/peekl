package endpoints

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/pkg/api/requests"
	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/environments"
	"github.com/peeklapp/peekl/pkg/models"
	"github.com/peeklapp/peekl/pkg/roles"
)

func PostRetrieveFile(ctx fiber.Ctx) error {
	var input requests.RetrieveFile
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body Invalid",
			Details: err.Error(),
		})
		return nil
	}

	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	if !roles.RoleNameIsValid(input.RoleName) {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Role name is not valid",
			Details: fmt.Sprintf("The provided role name (%s) is not valid.", input.RoleName),
		})
		return nil
	}

	if !environments.EnvironmentNameIsValid(input.Environment) {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Environment name is not valid",
			Details: fmt.Sprintf("The provided environment name (%s) is not valid.", input.Environment),
		})
	}

	envPath := filepath.Join(conf.Code.Directory, input.Environment)
	err := roles.DoesRoleExist(envPath, input.RoleName)
	if err != nil {
		if errors.As(err, &models.RoleNotFoundError{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Role could not be found",
				Details: err.Error(),
			})
			return nil
		} else {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: err.Error(),
			})
			return nil
		}
	}

	roleFilesPath := path.Join(envPath, "roles", input.RoleName, "files")
	root, err := os.OpenRoot(roleFilesPath)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
	}

	if _, err := root.Stat(input.Filename); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ctx.Status(400).JSON(responses.ErrorResponse{
				Error: "The file does not exist",
				Details: fmt.Sprintf(
					"The file '%s' could not be found inside of the role '%s' for environment '%s'",
					input.Filename,
					input.RoleName,
					input.Environment,
				),
			})
			return nil
		} else {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: err.Error(),
			})
			return nil
		}
	}

	fileContent, err := root.ReadFile(input.Filename)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
	}

	ctx.Status(200).JSON(responses.RetrieveFile{
		Filename: input.Filename,
		Content:  string(fileContent),
	})
	return nil
}

func PostRetrieveTemplate(ctx fiber.Ctx) error {
	var input requests.RetrieveTemplate
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body Invalid",
			Details: err.Error(),
		})
		return nil
	}

	conf, ok := ctx.Locals("config").(*config.ServerConfig)
	if !ok {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error: "Not OK !!!",
		})
	}

	if !roles.RoleNameIsValid(input.RoleName) {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Role name is not valid",
			Details: fmt.Sprintf("The provided role name (%s) is not valid.", input.RoleName),
		})
		return nil
	}

	if !environments.EnvironmentNameIsValid(input.Environment) {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Environment name is not valid",
			Details: fmt.Sprintf("The provided environment name (%s) is not valid.", input.Environment),
		})
	}

	envPath := filepath.Join(conf.Code.Directory, input.Environment)
	err := roles.DoesRoleExist(envPath, input.RoleName)
	if err != nil {
		if errors.As(err, &models.RoleNotFoundError{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Role could not be found",
				Details: err.Error(),
			})
			return nil
		} else {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: err.Error(),
			})
			return nil
		}
	}

	roleTemplatesPath := path.Join(envPath, "roles", input.RoleName, "templates")
	root, err := os.OpenRoot(roleTemplatesPath)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
	}

	var templateName string
	if strings.HasSuffix(input.TemplateName, ".tmpl") {
		templateName = input.TemplateName
	} else {
		templateName = fmt.Sprintf("%s.tmpl", input.TemplateName)
	}

	if _, err := root.Stat(templateName); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ctx.Status(400).JSON(responses.ErrorResponse{
				Error: "The template does not exist",
				Details: fmt.Sprintf(
					"The template '%s' could not be found inside of the role '%s' for environment '%s'",
					input.TemplateName,
					input.RoleName,
					input.Environment,
				),
			})
			return nil
		} else {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: err.Error(),
			})
			return nil
		}
	}

	fileContent, err := root.ReadFile(templateName)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
	}

	ctx.Status(200).JSON(responses.RetrieveFile{
		Filename: input.TemplateName,
		Content:  string(fileContent),
	})
	return nil
}
