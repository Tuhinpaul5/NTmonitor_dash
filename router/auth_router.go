package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterAuthRoutes handles all /api/auth/... endpoints
func RegisterAuthRoutes(api fiber.Router, h *handlers.AuthHandler) {
	auth := api.Group("/auth")

	auth.Post("/register", h.Register)
}
