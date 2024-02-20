package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func main() {
	// Create fiber app
	app := fiber.New(fiber.Config{})

	// Define routes
	v1 := app.Group("api/v1")
	v1.Post("hooks", HookCreatehook)
	v1.Post("hooks/:id", HookCreateHeartbeat)

	// Set Log Level
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	app.Listen(":3000")
}
