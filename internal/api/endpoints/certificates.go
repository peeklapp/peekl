package endpoints

import (
	"time"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/internal/api/requests"
	"github.com/peeklapp/peekl/internal/api/responses"
	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/database"
	"github.com/peeklapp/peekl/internal/utils"
)

func NewPostEnrollAgent(caData string, caKey string) (fiber.Handler, error) {
	loadedCa, err := certs.LoadCertificateFromData(caData)
	if err != nil {
		return nil, err
	}

	loadedKey, err := certs.LoadPKCS8PrivateKeyFromData(caKey)
	if err != nil {
		return nil, err
	}

	return func(ctx fiber.Ctx) error {
		var input requests.EnrollAgent
		if err := ctx.Bind().Body(&input); err != nil {
			return ctx.Status(400).JSON(responses.ErrorResponse{
				Error:   "Body invalid",
				Details: err.Error(),
			})
		}

		dbEngine, _ := ctx.Locals("databaseEngine").(*database.DatabaseEngine)

		// Get IP address from request
		ip := ctx.Req().IP()

		// Get token with IP address
		token, err := dbEngine.GetEnrollmentToken(ip)
		if err != nil {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Invalid token",
				Details: "The token could not be validated",
			})
		}

		// Check if provided token is valid
		valid, err := utils.VerifyPassword(input.Token, token.TokenHash)
		if err != nil {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Invalid token",
				Details: "The token could not be validated",
			})
		}

		if !valid {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Invalid token",
				Details: "The token could not be validated",
			})
		}

		// Check if token is still valid
		if time.Now().After(token.ValidUntil) {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Invalid token",
				Details: "The token could not be validated",
			})
		}

		// Validate the CSR
		loadedCsr, err := certs.LoadCertificateSigningRequest(input.CSR)
		if err != nil {
			return ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "An error happened while trying to load the sent certificate signing request",
			})
		}

		// Validate that we have at least one DNS name in CSR
		if len(loadedCsr.DNSNames) == 0 {
			return ctx.Status(400).JSON(responses.ErrorResponse{
				Error:   "Invalid Certificate Signing Request",
				Details: "The certificate signing request is not valid as it does not contain any DNS names",
			})
		}

		// Validate that vname in CSR is not already used
		used, err := dbEngine.IsNodeNameUsedInSigned(loadedCsr.DNSNames[0])
		if err != nil {
			return ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "Could not verify if the name in the certificate signing request is already used",
			})
		}
		if used {
			return ctx.Status(400).JSON(responses.ErrorResponse{
				Error:   "Node name is already used",
				Details: "The node name is already used, meaning that a certificate already exist for the given node name",
			})
		}

		// Sign the CSR and add to database
		signedCert, err := certs.SignCertificateSigningRequest(input.CSR, loadedCa, loadedKey)
		if err != nil {
			return ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "An error happened while trying to sign the certificate",
			})
		}

		// Insert signed certificate in database
		err = dbEngine.InsertSignedCertificate(loadedCsr.DNSNames[0], signedCert)
		if err != nil {
			return ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "An error happened while trying to insert signed certificate in database",
			})
		}

		// Delete the enrollment token as it has been used
		if err := dbEngine.DeleteEnrollmentToken(ip); err != nil {
			return ctx.Status(500).JSON(responses.ErrorResponse{
				Error:   "Internal Server Error",
				Details: "An error happened while trying to delete the enrollment token",
			})
		}

		// Reply with signed certificate and root CA
		return ctx.JSON(responses.EnrollAgent{
			CA:          caData,
			Certificate: signedCert,
		})
	}, nil
}
