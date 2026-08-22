package mtls

import (
	"crypto/x509"
	"time"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/internal/api/responses"
	"github.com/peeklapp/peekl/internal/certs"
	"github.com/peeklapp/peekl/internal/database"
)

func New(caPath string, databaseEngine *database.DatabaseEngine) (fiber.Handler, error) {
	// Load CA cert from path
	caCert, err := certs.LoadCertificateFromFile(caPath)
	if err != nil {
		return nil, err
	}

	// Create pool
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	// Create verify options
	verifyOptions := x509.VerifyOptions{
		Roots: caCertPool,
	}

	// Return actual handle
	return func(ctx fiber.Ctx) error {
		peerCertificates := ctx.RequestCtx().TLSConnectionState().PeerCertificates

		// Check if any certificate has been provided
		if len(peerCertificates) == 0 {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "No certificate provided",
				Details: "You have not provided any certificate with your request.",
			})
		}

		// Check if the certificate is signed by CA
		if _, err := peerCertificates[0].Verify(verifyOptions); err != nil {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Certificate invalid",
				Details: "The certificate that has been sent is not valid.",
			})
		}

		// Check if certificate is not expire
		if peerCertificates[0].NotAfter.Before(time.Now()) {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Certificate is expired",
				Details: "The certificate that has been set is expired.",
			})
		}

		// Check if the certificate is revoked
		if _, err := databaseEngine.GetRevokedCertificate(peerCertificates[0].SerialNumber.String()); err == nil {
			return ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Certificate has been revoked",
				Details: "The certificate that has been set is revoked.",
			})
		}

		// Let go through
		return ctx.Next()
	}, nil
}
