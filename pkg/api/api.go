package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	_ "github.com/redat00/peekl/docs"
	"github.com/redat00/peekl/pkg/api/endpoints"
	"github.com/redat00/peekl/pkg/certs"
	"github.com/redat00/peekl/pkg/config"
)

func NewApiEngine(conf *config.ServerConfig, certsDatabaseEngine *certs.CertsDatabaseEngine) (*fiber.App, error) {
	app := fiber.New()

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	// Attach configuration object to every request
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	// Create v1 group
	v1 := app.Group("v1")

	// Certificates group
	certGroup := v1.Group("certificates")

	// Attach certsDatabaseEngine object to every request
	certGroup.Use(func(c *fiber.Ctx) error {
		c.Locals("certsDatabaseEngine", certsDatabaseEngine)
		return c.Next()
	})
	certGroup.Get("/root", endpoints.GetRootCA)
	certGroup.Post("/submit", endpoints.PostSubmitCertificateRequest)
	certGroup.Post("/retrieve", endpoints.PostRetrieveSignedCertificate)

	return app, nil
}
