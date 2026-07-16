package main

import (
	"log"

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
func main() {
	cfg := config.LoadConfig()

	// Remove session middleware - we'll handle sessions manually
	conn := database.Connect(cfg.DBURL)
	mailer := services.New(cfg)

	// Repositories
	userRepo := repository.NewUserRepository(conn)
	otpRepo := repository.NewOtpRepository(conn)
	sessionRepo := repository.NewSessionRepository(conn)
	nodeRepo := repository.NewNodeRepository(conn)

	userHand := handlers.NewUserHandler(userRepo)
	authHand := handlers.NewAuthHandler(userRepo, otpRepo, sessionRepo, mailer, cfg)
	nodeHand := handlers.NewNodeHandler(nodeRepo)

	app := fiber.New(fiber.Config{
		AppName: "NTMonitor v1.0",
	})

	// Routes (no session middleware)
	router.SetupRoutes(app, userHand, authHand, nodeHand, sessionRepo, nodeRepo, cfg)

	log.Printf("NTMonitor is running on port %s", cfg.PORT)
	log.Fatal(app.Listen(":" + cfg.PORT))
}