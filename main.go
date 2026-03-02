package main

import (
	"log"
	"os"

	"NTMonitor/database"
	"NTMonitor/handlers"
	"NTMonitor/repository"
	"NTMonitor/router"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	conn := database.Connect(os.Getenv("DBURL"))

	userRepo := repository.NewUserRepository(conn)
	userHand := handlers.NewUserHandler(userRepo)

	app := fiber.New(fiber.Config{
		AppName: "NTMonitor v1.0",
	})

	router.SetupRoutes(app, userHand)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("NTMonitor is running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
