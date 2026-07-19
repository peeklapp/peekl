package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/peeklapp/peekl/internal/api/endpoints"
	"github.com/peeklapp/peekl/internal/api/middlewares/mtls"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/database"
	"github.com/peeklapp/peekl/internal/filecache"
)

func NewApiEngine(conf *config.ServerConfig, databaseEngine *database.DatabaseEngine) (*fiber.App, error) {
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
		c.Locals("databaseEngine", databaseEngine)
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
	catalogsGroup.Use(mtlsMiddleware)

	// -- Catalogs group needs access to server configuration
	catalogsGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	// -- Catalogs group endpoints
	catalogsGroup.Post("/catalog", endpoints.PostRetrieveCatalog)

	// Data group
	dataGroup := v1.Group("data")
	dataGroup.Use(mtlsMiddleware)

	// -- Create caches for both templates, and files
	templateCache := filecache.New()
	fileCache := filecache.New()

	// -- Data group needs access to server configuration
	dataGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	// -- Data group endpoints
	dataGroup.Post("/file", endpoints.NewPostRetrieveFileHandler(fileCache))
	dataGroup.Post("/template", endpoints.NewPostretrieveTemplate(templateCache))

	return app, nil
}
