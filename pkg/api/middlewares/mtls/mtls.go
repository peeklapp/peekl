package mtls

import (
	"crypto/x509"

	"github.com/gofiber/fiber/v3"
	"github.com/redat00/peekl/pkg/api/responses"
	"github.com/redat00/peekl/pkg/certs"
)

type MtlsMiddlewareConfig struct {
	CaCertificatePath string
}

func New(caPath string) (fiber.Handler, error) {
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
			ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "No certificate provided",
				Details: "You have not provided any certificate with your request.",
			})
			return nil
		}

		// Check if the certificate is valid
		if _, err := peerCertificates[0].Verify(verifyOptions); err != nil {
			ctx.Status(403).JSON(responses.ErrorResponse{
				Error:   "Certificate invalid",
				Details: "The certificate that has been sent is not valid.",
			})
			return nil
		}

		// TODO: IMPLEMENT CRL ?

		// Let go through
		return ctx.Next()
	}, nil
}
