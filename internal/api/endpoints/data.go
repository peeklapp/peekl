package endpoints

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/internal/api/responses"
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/utils"
)

func isAFolderMissing(ctx fiber.Ctx, dataRoot *os.Root, environment string, id string) bool {
	if !utils.FileExist(environment, dataRoot) {
		ctx.Status(404).JSON(responses.ErrorResponse{
			Error:   "Environment not found",
			Details: fmt.Sprintf("The environment '%s' could not be found inside the code directory", environment),
		})
		return true
	}

	if !utils.FileExist(filepath.Join(environment, id), dataRoot) {
		ctx.Status(404).JSON(responses.ErrorResponse{
			Error:   "ID not found",
			Details: fmt.Sprintf("The ID '%s' could not be found inside the code directory for environment '%s'", id, environment),
		})
		return true
	}

	return false
}

func NewGetNodeData(dataRoot *os.Root) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		conf, _ := ctx.Locals("config").(*config.ServerConfig)
		environment := filepath.Base(ctx.Params("environment"))
		id := filepath.Base(ctx.Params("id"))
		tarballName := filepath.Base(ctx.Params("name"))

		certNodeName := ctx.RequestCtx().TLSConnectionState().PeerCertificates[0].Subject.CommonName
		if tarballName != fmt.Sprintf("%s%s", certNodeName, code.TarballExtension) {
			ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Forbidden",
				Details: "The name found in the certificate is not the same as the one set in the request",
			})
			return nil
		}

		if isAFolderMissing(ctx, dataRoot, environment, id) {
			return nil
		}

		if !utils.FileExist(filepath.Join(environment, id, "nodes", tarballName), dataRoot) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Node tarball not found",
				Details: fmt.Sprintf("The node tarball '%s' could not be found inside the code directory for ID '%s' and environment '%s'", tarballName, id, environment),
			})
			return nil
		}

		return ctx.SendFile(filepath.Join(conf.Code.Directory, environment, id, "nodes", tarballName))
	}
}

func NewGetCodeData(dataRoot *os.Root) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		conf, _ := ctx.Locals("config").(*config.ServerConfig)
		environment := filepath.Base(ctx.Params("environment"))
		id := filepath.Base(ctx.Params("id"))

		if isAFolderMissing(ctx, dataRoot, environment, id) {
			return nil
		}

		if !utils.FileExist(filepath.Join(environment, id, code.CodeTarballName), dataRoot) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Node tarball not found",
				Details: fmt.Sprintf("The code tarball '%s' could not be found inside the code directory for ID '%s' and environment '%s'", code.CodeTarballName, id, environment),
			})
			return nil
		}

		return ctx.SendFile(filepath.Join(conf.Code.Directory, environment, id, code.CodeTarballName))
	}
}
