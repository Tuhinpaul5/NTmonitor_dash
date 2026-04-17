package middleware

import (
	"NTMonitor/config"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTMiddleware validates JWT tokens from Authorization header or query parameter
func JWTMiddleware(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get token from Authorization header or query parameter
		tokenString := extractToken(c)
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or invalid authorization token",
			})
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid signing method")
			}
			return []byte(cfg.JWT_SECRET), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Store token in context for handler access
		c.Locals("user", token)

		return c.Next()
	}
}

// extractToken gets the JWT token from Authorization header or query parameter
func extractToken(c fiber.Ctx) string {
	// Try Authorization header first (Bearer token)
	auth := c.Get("Authorization")
	if auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try query parameter as fallback
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}
