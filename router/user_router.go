package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterUserRoutes handles all /api/v1/users/... endpoints
func RegisterUserRoutes(v1 fiber.Router, h *handlers.UserHandler) {
	users := v1.Group("/users")

	users.Post("/", h.CreateUser)
	users.Get("/", h.GetUsers)
	users.Get("/id/:id", h.GetUser)
}
