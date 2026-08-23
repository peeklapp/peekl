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
	v1 := app.Group("v1")

	log := logrus.New()
	loggerMiddleware := logger.NewLogger(log)
	app.Use(loggerMiddleware)

	// Create mTLS middleware
	mtlsMiddleware, err := mtls.New(conf.Certificates.CaCertificateFilePath, databaseEngine)
	if err != nil {
		return nil, err
	}

	// Get raw CA certificate
	caCert, err := os.ReadFile(conf.Certificates.CaCertificateFilePath)
	if err != nil {
		return nil, err
	}

	// Get raw CA key
	caKey, err := os.ReadFile(conf.Certificates.CaCertificateKeyPath)
	if err != nil {
		return nil, err
	}

	// Certificates group
	certificatesGroup := v1.Group("certificates")
	certificatesGroup.Use(func(c fiber.Ctx) error {
		c.Locals("databaseEngine", databaseEngine)
		return c.Next()
	})
	certificatesGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})
	enrollEndpoint, err := endpoints.NewPostEnrollAgent(string(caCert), string(caKey))
	if err != nil {
		return nil, err
	}
	certificatesGroup.Post("/enroll", enrollEndpoint)

	// Create data root
	dataRoot, err := os.OpenRoot(conf.Code.Directory)
	if err != nil {
		return nil, err
	}

	// Catalogs group
	catalogsGroup := v1.Group("catalogs")
	catalogsGroup.Use(mtlsMiddleware)
	catalogsFilecache := filecache.New()
	catalogsGroup.Post("/catalog", endpoints.NewPostGetCatalog(dataRoot, catalogsFilecache))

	// Data group
	dataGroup := v1.Group("data")
	dataGroup.Use(mtlsMiddleware)
	dataGroup.Use(func(c fiber.Ctx) error {
		c.Locals("config", conf)
		return c.Next()
	})
	dataGroup.Get("/:environment/:id/nodes/:name", endpoints.NewGetNodeData(dataRoot))
	dataGroup.Get("/:environment/:id/"+code.CodeTarballName, endpoints.NewGetCodeData(dataRoot))

	return app, nil
}
