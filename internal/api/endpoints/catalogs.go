package endpoints

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/internal/api/requests"
	"github.com/peeklapp/peekl/internal/api/responses"
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/environments"
	"github.com/peeklapp/peekl/internal/filecache"
)

func NewPostInquiryForCatalog(dataRoot *os.Root, filecache *filecache.FileCache) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		var input requests.InquiryForCatalog
		if err := ctx.Bind().Body(&input); err != nil {
			ctx.Status(400).JSON(responses.ErrorResponse{
				Error:   "Body Invalid",
				Details: err.Error(),
			})
			return nil
		}

		if !environments.EnvironmentNameIsValid(input.Environment) {
			ctx.Status(400).JSON(responses.ErrorResponse{
				Error:   "Environment name is not valid",
				Details: fmt.Sprintf("The provided environment name %s is not valid", input.Environment),
			})
			return nil
		}

		if _, err := dataRoot.Stat(input.Environment); os.IsNotExist(err) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "Environment not found",
				Details: "The request environment could not be fond on the server",
			})
			return nil
		}

		latestId, err := code.GetLatestVersionInEnvironment(dataRoot, input.Environment)
		if err != nil {
			ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "An issue happened while trying to find the latest ID for environment",
			})
			return nil
		}

		nodeName := ctx.RequestCtx().TLSConnectionState().PeerCertificates[0].Subject.CommonName
		nodeFileInfo, err := filecache.GetInfo(dataRoot, filepath.Join(input.Environment, latestId, "nodes", nodeName+code.TarballExtension))
		if err != nil {
			if os.IsNotExist(err) {
				ctx.Status(404).JSON(responses.ErrorResponse{
					Error:   "Node not found in environment",
					Details: fmt.Sprintf("The node (%s) could not be found in provided environment (%s)", nodeName, input.Environment),
				})
				return nil
			} else {
				ctx.Status(500).JSON(responses.ErrorResponse{
					Error:   "Internal Server Error",
					Details: "An issue happened while trying to gather node tarball information",
				})
				return nil
			}
		}

		codeFileInfo, err := filecache.GetInfo(dataRoot, filepath.Join(input.Environment, latestId, code.CodeTarballName))
		if err != nil {
			if os.IsNotExist(err) {
				ctx.Status(404).JSON(responses.ErrorResponse{
					Error:   "Code file not found on server",
					Details: fmt.Sprintf("The code file tarball could not be found in provided environment (%s)", input.Environment),
				})
				return nil
			} else {
				ctx.Status(500).JSON(responses.ErrorResponse{
					Error:   "Internal Server Error",
					Details: "An issue happened while trying to gather code tarball information",
				})
				return nil
			}
		}

		ctx.Status(200).JSON(responses.InquiryForCatalog{
			NodeTarball: responses.FileResponseEntry{
				Path: fmt.Sprintf("v1/data/%s/%s/nodes/%s%s", input.Environment, latestId, nodeName, code.TarballExtension),
				Hash: nodeFileInfo.Hash,
			},
			CodeTarball: responses.FileResponseEntry{
				Path: fmt.Sprintf("v1/data/%s/%s/%s", input.Environment, latestId, code.CodeTarballName),
				Hash: codeFileInfo.Hash,
			},
		})
		return nil
	}
}
