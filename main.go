package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

type LoggerMiddlewareConfig struct {
	Next func(ctx *fiber.Ctx) bool
}

func LoggerMiddleware(conf ...LoggerMiddlewareConfig) fiber.Handler {
	zLogger := log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	var config LoggerMiddlewareConfig
	if len(conf) > 0 {
		config = conf[0]
	}
	return func(fiberCtx fiber.Ctx) error {
		if config.Next != nil && config.Next(&fiberCtx) {
			return fiberCtx.Next()
		}

		startTime := time.Now()

		statusCode := fiberCtx.Response().StatusCode()
		returnedLogger := zLogger.With().
			Int("status", statusCode).
			Str("method", fiberCtx.Method()).
			Str("path", fiberCtx.Path()).
			Str("ip", fiberCtx.IP()).
			Str("duration", time.Since(startTime).String()).
			Str("user-agent", fiberCtx.Get(fiber.HeaderUserAgent)).
			Logger()

		msg := "Request: "
		if statusCode >= fiber.StatusBadRequest && statusCode < fiber.StatusInternalServerError {
			returnedLogger.Warn().Msg(msg)
		} else if statusCode >= fiber.StatusInternalServerError {
			returnedLogger.Error().Msg(msg)
		} else {
			returnedLogger.Info().Msg(msg)
		}
		return nil
	}
}

func main() {
	// Set Log Level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create fiber app
	app := fiber.New(fiber.Config{})
	app.Use(requestid.New())
	app.Use(LoggerMiddleware())

	// Define routes
	v1 := app.Group("api/v1")
	v1.Post("hooks", HookCreatehook)
	v1.Post("hooks/:id", HookCreateHeartbeat)

	app.Listen(":3000")
}
