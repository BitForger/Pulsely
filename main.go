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
	Next   func(ctx *fiber.Ctx) bool
	Logger *zerolog.Logger
}

func LoggerMiddleware(conf ...LoggerMiddlewareConfig) fiber.Handler {
	var zLogger zerolog.Logger
	var config LoggerMiddlewareConfig
	if len(conf) > 0 {
		config = conf[0]
	}

	if config.Logger != nil {
		zLogger = *config.Logger
	} else {
		zLogger = zerolog.New(zerolog.NewConsoleWriter())
	}

	return func(fiberCtx fiber.Ctx) error {
		if config.Next != nil && config.Next(&fiberCtx) {
			log.Debug().Msg("Going to next")
			return fiberCtx.Next()
		}

		startTime := time.Now()

		statusCode := fiberCtx.Response().StatusCode()
		returnedLogger := zLogger.With().
			Int("status", statusCode).
			Str("method", fiberCtx.Method()).
			Str("path", fiberCtx.Path()).
			Str("ip", fiberCtx.IP()).
			// Note: doesn't actually calc duration of request
			Str("duration", time.Since(startTime).String()).
			Str("user-agent", fiberCtx.Get(fiber.HeaderUserAgent)).
			Logger()

		msg := "Request:"
		if err := fiberCtx.Context().Err(); err != nil {
			msg = err.Error()
		}
		if statusCode >= fiber.StatusBadRequest && statusCode < fiber.StatusInternalServerError {
			returnedLogger.Warn().Msg(msg)
		} else if statusCode >= fiber.StatusInternalServerError {
			returnedLogger.Error().Msg(msg)
		} else {
			returnedLogger.Info().Msg(msg)
		}
		return fiberCtx.Next()
	}
}

func main() {
	environment, lookupOk := os.LookupEnv("GO_ENV")
	if !lookupOk {
		log.Warn().Str("environment", environment).Msg("failed to lookup go env")
	}
	if environment == "" {
		environment = "production"
	}

	// Set Log Level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if environment != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Create fiber app
	app := fiber.New(fiber.Config{})
	app.Use(requestid.New())
	app.Use(LoggerMiddleware(LoggerMiddlewareConfig{Logger: &log.Logger}))

	// Define routes
	v1 := app.Group("api/v1")
	v1.Post("hooks", HookCreateHook)
	v1.Post("hooks/:id", HookCreateHeartbeat)
	v1.Patch("hooks/:id", UpdateHeartbeat)

	app.Listen(":3000")
}
