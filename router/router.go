package router

import (
	"NTMonitor/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App, userHand *handlers.UserHandler) {
	// Global Middleware
	app.Use(logger.New())

	// Public Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Versioned Group
	v1 := app.Group("/api")

	RegisterUserRoutes(v1, userHand)
}