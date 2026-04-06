package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v3"
)

// RegisterUserRoutes handles all /api/users/... endpoints (public routes)
func RegisterUserRoutes(app fiber.Router, h *handlers.UserHandler) {
	users := app.Group("/users")

	// Public routes (no auth required)
	users.Post("/", h.CreateUser) // If you want user creation to be public
}

// RegisterProtectedUserRoutes handles protected /api/users/... endpoints
func RegisterProtectedUserRoutes(app fiber.Router, h *handlers.UserHandler, authMiddleware fiber.Handler) {
	users := app.Group("/users")
	users.Use(authMiddleware) // Apply auth middleware to all user routes

	// Protected routes (auth required)
	users.Get("/", h.GetUsers)
	users.Get("/:id", h.GetUser)
}
