package router

import (
	_ "NTMonitor/docs" // <--- CRITICAL: Import your generated docs
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"github.com/gofiber/fiber/v2/middleware/cors" // Add this
)

func SetupRoutes(
	app *fiber.App,
	userHand *handlers.UserHandler,
	authHand *handlers.AuthHandler,
) {

	
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*", // Allow all origins for development
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	app.Use(logger.New())

	// 1. Serve the generated Swagger JSON from the /docs folder
	// This makes http://localhost:8000/swagger/swagger.json accessible
	app.Static("/swagger", "./docs")

	// 2. Scalar Documentation UI
	app.Get("/docs", func(c *fiber.Ctx) error {
		html := `
		<!doctype html>
		<html>
		  <head>
			<title>NTMonitor API Reference</title>
			<meta charset="utf-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1" />
		  </head>
		  <body>
			<script id="api-reference" data-url="/swagger/swagger.json"></script>
			<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
		  </body>
		</html>`

		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	// Public Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Versioned Group
	api := app.Group("/api")
	RegisterUserRoutes(api, userHand)
	RegisterAuthRoutes(api, authHand)
}
