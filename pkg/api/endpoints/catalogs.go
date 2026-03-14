package endpoints

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/pkg/api/requests"
	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/catalog"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/environments"
	"github.com/peeklapp/peekl/pkg/models"
)

func PostRetrieveCatalog(ctx fiber.Ctx) error {
	var input requests.RetrieveCatalog
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body Invalid",
			Details: err.Error(),
		})
		return nil
	}

	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	if !environments.EnvironmentNameIsValid(input.Environment) {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Environment name is not valid",
			Details: fmt.Sprintf("The provided environment name %s is not valid", input.Environment),
		})
		return nil
	}

	directoryPath := path.Join(conf.Code.Directory, input.Environment)
	nodeName := ctx.RequestCtx().TLSConnectionState().PeerCertificates[0].Subject.CommonName

	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		ctx.Status(404).JSON(responses.ErrorResponse{
			Error: "Environment not found in code folder.",
			Details: models.EnvironmentNotFoundError{
				Environment: input.Environment,
			}.Error(),
		})
		return nil
	}

	resources, loadedRoles, tags, variables, err := catalog.CompileCatalog(directoryPath, nodeName)
	if err != nil {
		if errors.As(err, &models.NodeNotFoundError{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Node not found in inventory",
				Details: err.Error(),
			})
			return nil
		} else if errors.As(err, &models.GroupNotFoundError{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Group not found in inventory",
				Details: err.Error(),
			})
			return nil
		} else if errors.As(err, &models.RoleNotFoundError{}) {
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
		}
	}

	ctx.Status(200).JSON(
		responses.GetCatalog{
			GlobalResource: resources,
			Roles:          loadedRoles,
			Tags:           tags,
			Variables:      variables,
		},
	)
	return nil
}
