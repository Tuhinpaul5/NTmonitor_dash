package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes handles all /api/users/... endpoints
func RegisterUserRoutes(app fiber.Router, h *handlers.UserHandler) {
	users := app.Group("/users")

	users.Post("/", h.CreateUser)
	users.Get("/", h.GetUsers)
	users.Get("/id/:id", h.GetUser)
}
