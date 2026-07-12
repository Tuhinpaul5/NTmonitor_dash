package middleware

import (
	"NTMonitor/repository"
	"log"

	"github.com/gofiber/fiber/v3"
)

// AuthMiddleware checks for valid session tokens
func AuthMiddleware(sessionRepo *repository.SessionRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get session token from cookie
		sessionToken := c.Cookies("session_token")
		if sessionToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Check if session exists and is valid
		session, err := sessionRepo.FindByToken(sessionToken)
		if err != nil {
			log.Printf("Invalid session token: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Store user ID in context for use in handlers
		c.Locals("user_id", session.UserID)

		return c.Next()
	}
}
