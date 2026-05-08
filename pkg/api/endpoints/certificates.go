package endpoints

import (
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/pkg/api/requests"
	"github.com/peeklapp/peekl/pkg/api/responses"
	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
	"github.com/peeklapp/peekl/pkg/database"
	"github.com/peeklapp/peekl/pkg/models"
)

// This file contains all the API routes related to certificates

func GetRootCA(ctx fiber.Ctx) error {
	// Get configuration from context
	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	// Get local CA file
	res, err := os.ReadFile(conf.Certificates.CaCertificateFilePath)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return nil
	}

	// Return CA
	ctx.Status(200).JSON(responses.GetRootCA{
		Certificate: string(res),
	})
	return nil
}

func PostSubmitCertificateRequest(ctx fiber.Ctx) error {
	var input requests.SubmitCertificateRequest
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body invalid",
			Details: err.Error(),
		})
		return nil
	}

	dbEngine, _ := ctx.Locals("databaseEngine").(*database.DatabaseEngine)

	loadedCsr, err := certs.LoadCertificateSigningRequest(input.CSR)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return nil
	}

	nodeNameUsed, err := dbEngine.IsNodeNameUsed(loadedCsr.DNSNames[0])
	if err != nil {
		println(loadedCsr.DNSNames[0])
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return nil
	}

	if nodeNameUsed {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error: "Node name already used",
			Details: fmt.Sprintf(
				"Node name `%s` cannot be used to generate a new certificate, as a similar certificate already exist.",
				loadedCsr.DNSNames[0],
			),
		})
		return nil
	}

	err = dbEngine.InsertPendingCertificate(loadedCsr.DNSNames[0], input.CSR)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return nil
	}

	ctx.Status(201).JSON(responses.MessageResponse{
		Details: "CSR submitted to the server.",
	})
	return nil
}

func PostRetrieveSignedCertificate(ctx fiber.Ctx) error {
	var input requests.RetrieveSignedCertificate
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body invalid",
			Details: err.Error(),
		})
		return nil
	}

	dbEngine, _ := ctx.Locals("databaseEngine").(*database.DatabaseEngine)

	signedCert, err := dbEngine.GetSignedCertificate(input.CsrSignature)
	if err != nil {
		if errors.Is(err, models.SignedCertificateNotFound{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "No signed certificate correspond to given CSR signature",
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

	ctx.Status(200).JSON(responses.RetrieveSignedCertificate{
		Certificate: signedCert.Data,
	})
	return nil
}
