package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/peeklapp/peekl/pkg/api/endpoints"
	"github.com/peeklapp/peekl/pkg/api/middlewares/mtls"
	"github.com/peeklapp/peekl/pkg/certs"
	"github.com/peeklapp/peekl/pkg/config"
)

func NewApiEngine(conf *config.ServerConfig, certsDatabaseEngine *certs.CertsDatabaseEngine) (*fiber.App, error) {
	// Create app instance
	app := fiber.New()

	loggerMiddleware := logger.New()
	app.Use(loggerMiddleware)

	// Create mTLS middleware
	mtlsMiddleware, err := mtls.New(conf.Certificates.CaCertificateFilePath)
	if err != nil {
		return nil, err
	}

	// Create v1 group
	v1 := app.Group("v1")

	// Certificates group
	certificatesGroup := v1.Group("certificates")

	// -- Certificates group needs access to certificate database engine
	certificatesGroup.Use(func(c fiber.Ctx) error {
		c.Locals("certsDatabaseEngine", certsDatabaseEngine)
		return c.Next()
	})

	// -- Certificates group needs access to server configuration
	certificatesGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	// -- Certificates group endpoints
	certificatesGroup.Get("/root", endpoints.GetRootCA)
	certificatesGroup.Post("/submit", endpoints.PostSubmitCertificateRequest)
	certificatesGroup.Post("/retrieve", endpoints.PostRetrieveSignedCertificate)

	// Catalogs group
	catalogsGroup := v1.Group("catalogs")

	// -- Catalogs group needs access to server configuration
	catalogsGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	// -- Catalogs group endpoints
	catalogsGroup.Use(mtlsMiddleware)
	catalogsGroup.Get("/catalog", endpoints.GetCatalog)

	return app, nil
}
