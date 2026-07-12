package router

import (
	"NTMonitor/handlers"
	"NTMonitor/middleware"
	"NTMonitor/models"
	"NTMonitor/repository"

	"github.com/gofiber/fiber/v3"
)

// // RegisterUserRoutes handles all /api/users/... endpoints (public routes)
// func RegisterUserRoutes(app fiber.Router, h *handlers.UserHandler) {
// 	users := app.Group("/users")

// 	// Public routes (no auth required)
// 	users.Post("/", h.CreateUser) // If you want user creation to be public
// }

// RegisterProtectedUserRoutes handles protected /api/users/... endpoints
func RegisterProtectedUserRoutes(app fiber.Router, h *handlers.UserHandler, authMiddleware fiber.Handler, userRepo *repository.UserRepository) {
	users := app.Group("/users")
	users.Use(authMiddleware) // Apply auth middleware to all user routes

	users.Get("/get-me", h.GetMe) // Get current user info
	// Admin only - Get all users
	users.Get("/", middleware.RequireRole(userRepo, models.UserTypeAdmin), h.GetUsers)

	// Any authenticated user - Get specific user (ownership check in handler)
	users.Get("/:id", h.GetUser)
}