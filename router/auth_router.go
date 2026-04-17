package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v3"
)

// RegisterAuthRoutes handles all /api/auth/... endpoints
func RegisterAuthRoutes(api fiber.Router, h *handlers.AuthHandler, jwtMiddleware fiber.Handler) {
	auth := api.Group("/auth")

	// Public routes (no middleware)
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Post("/send-otp", h.SendOTP)
	auth.Post("/verify-otp", h.VerifyOTP)

	// Protected route (requires JWT from OTP verification)
	auth.Post("/register", jwtMiddleware, h.Register)
}
