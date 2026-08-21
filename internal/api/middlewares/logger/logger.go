package logger

import (
	"time"

	fiber "github.com/gofiber/fiber/v3"
	"github.com/sirupsen/logrus"
)

func NewLogger(log *logrus.Logger) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()

		err := ctx.Next()

		status := ctx.Response().StatusCode()

		entry := log.WithFields(logrus.Fields{
			"method":  ctx.Method(),
			"path":    ctx.Path(),
			"status":  status,
			"latency": time.Since(start).String(),
			"ip":      ctx.IP(),
		})

		if err != nil {
			entry = entry.WithField("error", err.Error())
		}

		switch {
		case status >= 500:
			entry.Error("server error")
		case status >= 400:
			entry.Warn("client error")
		default:
			entry.Info("request handled")
		}

		return err
	}
}
