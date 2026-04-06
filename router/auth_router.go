package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v3"
)

// RegisterAuthRoutes handles all /api/auth/... endpoints
func RegisterAuthRoutes(api fiber.Router, h *handlers.AuthHandler) {
	auth := api.Group("/auth")

	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/send-otp", h.SendOTP)
	auth.Post("/verify-otp", h.VerifyOTP)
}
