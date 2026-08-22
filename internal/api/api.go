package api

import (
	"os"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/peeklapp/peekl/internal/api/endpoints"
	"github.com/peeklapp/peekl/internal/api/middlewares/logger"
	"github.com/peeklapp/peekl/internal/api/middlewares/mtls"
	"github.com/peeklapp/peekl/internal/code"
	"github.com/peeklapp/peekl/internal/config"
	"github.com/peeklapp/peekl/internal/database"
	"github.com/peeklapp/peekl/internal/filecache"
	"github.com/sirupsen/logrus"
)

func NewApiEngine(conf *config.ServerConfig, databaseEngine *database.DatabaseEngine) (*fiber.App, error) {
	// Create app instance
	app := fiber.New()

	log := logrus.New()
	loggerMiddleware := logger.NewLogger(log)
	app.Use(loggerMiddleware)

	// Create mTLS middleware
	mtlsMiddleware, err := mtls.New(conf.Certificates.CaCertificateFilePath, databaseEngine)
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

	// Create data root
	dataRoot, err := os.OpenRoot(conf.Code.Directory)
	if err != nil {
		return nil, err
	}

	// Catalogs group
	catalogsGroup := v1.Group("catalogs")
	catalogsGroup.Use(mtlsMiddleware)
	catalogsFilecache := filecache.New()

	// -- Catalogs group endpoints
	catalogsGroup.Post("/catalog", endpoints.NewPostGetCatalog(dataRoot, catalogsFilecache))

	// Data group
	dataGroup := v1.Group("data")

	// -- Data group needs access to server configurtion
	dataGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})

	dataGroup.Get("/:environment/:id/nodes/:name", endpoints.NewGetNodeData(dataRoot))
	dataGroup.Get("/:environment/:id/"+code.CodeTarballName, endpoints.NewGetCodeData(dataRoot))

	return app, nil
}
