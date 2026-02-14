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
)

// This file contains all the API routes related to certificates

// GetRootCA godoc
// @Summary     Used to get the root CA from the server
// @Description get root ca from server
// @Tags        certificates
// @Produce     json
// @Success     200 {object} responses.GetRootCA
// @Router      /v1/certificates/root [get]
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

// SubmitCertificate godoc
// @Summary     Used to submit a certificate
// @Description submit an unsigned certificate
// @Tags        certificates
// @Accept      json
// @Produce     json
// @Param       data body requests.SubmitCertificateRequest true "Name of the node, along the CSR"
// @Success     201 {object} responses.MessageResponse
// @Failure     400 {object} responses.ErrorResponse
// @Router      /v1/certificates/submit [post]
func PostSubmitCertificateRequest(ctx fiber.Ctx) error {
	var input requests.SubmitCertificateRequest
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body invalid",
			Details: err.Error(),
		})
		return nil
	}

	// TODO: ADD VALIDATION TO THE BODY

	// Get configuration
	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	// Get certsDatabaseEngine
	certsDbEngine, _ := ctx.Locals("certsDatabaseEngine").(*certs.CertsDatabaseEngine)

	// Create local file containing CSR
	filePath := fmt.Sprintf("%s/%s.csr", conf.Certificates.PendingDirectory, input.NodeName)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	file.Write([]byte(input.CSR))

	// Get CSR signature
	csrSignature := certs.GetCertificateSigningRequestSignature(input.CSR)

	// Insert Pending certificate in database
	certsDbEngine.InsertPendingCertificate(input.NodeName, csrSignature)

	// Send succesful answer
	ctx.Status(201).JSON(responses.MessageResponse{
		Details: "CSR submitted to the server.",
	})
	return nil
}

// GetSignedCertificate godoc
// @Summary     Retrieve a certificate once it has been signed
// @Description retrieve a node signed certificate
// @Tags        certificates
// @Accept      json
// @Produce     json
// @Param       data body requests.RetrieveSignedCertificate true "Name of the node, along the CSR"
// @Success     200 {object} responses.RetrieveSignedCertificate
// @Failure     400 {object} responses.ErrorResponse
// @Failure     404 {object} responses.ErrorResponse
// @Router      /v1/certificates/retrieve [post]
func PostRetrieveSignedCertificate(ctx fiber.Ctx) error {
	var input requests.RetrieveSignedCertificate
	if err := ctx.Bind().Body(&input); err != nil {
		ctx.Status(400).JSON(responses.ErrorResponse{
			Error:   "Body invalid",
			Details: err.Error(),
		})
		return nil
	}

	// TODO: ADD VALIDATION OF THE BODY

	// Get configuration
	conf, _ := ctx.Locals("config").(*config.ServerConfig)

	// Get certsDatabaseEngine
	certsDbEngine, _ := ctx.Locals("certsDatabaseEngine").(*certs.CertsDatabaseEngine)

	// Get CSR signature
	csrSignature := certs.GetCertificateSigningRequestSignature(input.CSR)

	// Get node name from CSR
	signedCertDb, err := certsDbEngine.GetSignedCertificate(input.NodeName)
	if err != nil {
		if errors.Is(err, certs.SignedCertificateNotFound{}) {
			ctx.Status(404).JSON(responses.ErrorResponse{
				Error:   "No signed certificate correspond to given node name",
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

	// Make sure that CSR correspond to the one we have
	if csrSignature != signedCertDb.CsrSignature {
		ctx.Status(403).JSON(responses.ErrorResponse{
			Error:   "CSR does not match",
			Details: "The CSR that has been provided does not match the one found server-side",
		})
		return nil
	}

	// Send back the signed certificate
	signedCert, err := os.ReadFile(
		fmt.Sprintf(
			"%s/%s.pem",
			conf.Certificates.SignedDirectory,
			input.NodeName,
		),
	)
	if err != nil {
		ctx.Status(500).JSON(responses.ErrorResponse{
			Error:   "Internal Server Error",
			Details: err.Error(),
		})
		return nil
	}

	ctx.Status(200).JSON(responses.RetrieveSignedCertificate{
		Certificate: string(signedCert),
	})
	return nil
}
