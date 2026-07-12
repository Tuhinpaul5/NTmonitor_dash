package main

import (
	"log"
	"time"

	"NTMonitor/config"
	"NTMonitor/database"
	"NTMonitor/handlers"
	"NTMonitor/repository"
	"NTMonitor/router"
	"NTMonitor/services"

	"github.com/gofiber/fiber/v3"
)

//	@title			NTMonitor API
//	@version		1.0
//	@description	API documentation for NTMonitor application
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@ntmonitor.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8000
//	@BasePath	/

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey	SessionAuth
// @in							cookie
// @name						session_token
// @description				Session token for authenticated requests.

func main() {
	cfg := config.LoadConfig()

	conn := database.Connect(cfg.DBURL)
	mailer := services.New(cfg)

	// Repositories
	userRepo := repository.NewUserRepository(conn)
	otpRepo := repository.NewOtpRepository(conn)
	sessionRepo := repository.NewSessionRepository(conn)

	userHand := handlers.NewUserHandler(userRepo)
	authHand := handlers.NewAuthHandler(userRepo, otpRepo, sessionRepo, mailer, cfg)

	// Initialize WebSocket Gateway and Command handlers
	registry := handlers.NewConnectionRegistry()
	gatewayHandler := handlers.NewGatewayWSHandler(registry)
	commandHandler := handlers.NewCommandHandler(registry)

	// Start connection health monitor
	// Ping every 30s, cleanup every 60s, max 3 missed pings
	gatewayHandler.StartHealthMonitor(30*time.Second, 60*time.Second, 3)

	app := fiber.New(fiber.Config{
		AppName: "NTMonitor v1.0",
	})

	// Routes (no session middleware)
	router.SetupRoutes(app, userHand, authHand, sessionRepo, userRepo, cfg, gatewayHandler, commandHandler)

	log.Printf("NTMonitor is running on port %s", cfg.PORT)
	log.Fatal(app.Listen(":" + cfg.PORT))
}
