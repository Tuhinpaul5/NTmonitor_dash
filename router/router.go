package router

import (
	"NTMonitor/config"
	_ "NTMonitor/docs" // <--- CRITICAL: Import your generated docs
	"NTMonitor/handlers"
	"NTMonitor/middleware"
	"NTMonitor/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/gofiber/fiber/v3/middleware/cors" // Add this
)

func SetupRoutes(
	app *fiber.App,
	userHand *handlers.UserHandler,
	authHand *handlers.AuthHandler,
	sessionRepo *repository.SessionRepository,
	userRepo *repository.UserRepository,
	cfg *config.Config,
	gatewayHandler *handlers.GatewayWSHandler,
	commandHandler *handlers.CommandHandler,
) {

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000", "http://127.0.0.1:8000"}, // Allow your frontend origin
		AllowHeaders:     []string{"*"},                                              // Allow all headers
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"*"}, // Expose all headers
	}))

	app.Use(logger.New())

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Howdy!! 🤠")
	})

	// 1. Serve the generated Swagger JSON from the /docs folder
	// This makes http://localhost:8000/swagger/swagger.json accessible
	app.Get("/swagger/swagger.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendFile("./docs/swagger.json")
	})

	app.Get("/swagger/*", func(c fiber.Ctx) error {
		// Extract the file path from the URL
		filePath := c.Params("*")
		if filePath == "" {
			filePath = "swagger.json"
		}

		// Serve files from the docs directory
		return c.SendFile("./docs/" + filePath)
	})

	// 2. API Documentation UI (with fallback options)
	app.Get("/docs", func(c fiber.Ctx) error {
		html := `<!doctype html>
<html>
  <head>
    <title>NTMonitor API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="scalar-container"></div>
    <script 
      id="api-reference" 
      data-url="/swagger/swagger.json"
      data-configuration='{"theme":"purple","hideDownloadButton":true}'
    ></script>
    <script>
      // Fallback error handling
      window.addEventListener('error', function(e) {
        console.error('Error loading Scalar:', e);
        document.getElementById('scalar-container').innerHTML = 
          '<h1>API Documentation</h1>' +
          '<p>Swagger JSON available at: <a href="/swagger/swagger.json">/swagger/swagger.json</a></p>' +
          '<p>Alternative Swagger UI: <a href="/swagger-ui">Swagger UI</a></p>';
      });
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// 3. Alternative Swagger UI
	app.Get("/swagger-ui", func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
  <head>
    <title>NTMonitor API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
      SwaggerUIBundle({
        url: '/swagger/swagger.json',
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.presets.standalone
        ]
      });
    </script>
  </body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// Public Routes
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Versioned Group
	api := app.Group("/api")

	// Middleware
	jwtMiddleware := middleware.JWTMiddleware(cfg)
	sessionAuthMiddleware := middleware.AuthMiddleware(sessionRepo)

	// Public routes (no authentication required)
	RegisterAuthRoutes(api, authHand, jwtMiddleware)
	// RegisterUserRoutes(api, userHand) // Public user routes if any

	// Protected routes (require session authentication)
	RegisterProtectedUserRoutes(api, userHand, sessionAuthMiddleware, userRepo)

	// Gateway WebSocket and Command routes
	RegisterGatewayRoutes(api, gatewayHandler, commandHandler, jwtMiddleware)
	RegisterCommandRoutes(api, commandHandler, sessionAuthMiddleware)
}
