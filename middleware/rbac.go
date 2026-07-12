package middleware

import (
	"NTMonitor/models"
	"NTMonitor/repository"

	"github.com/gofiber/fiber/v3"
)

// RequireRole creates a middleware that checks if the authenticated user has one of the allowed roles
// Usage:
//   - Single role: RequireRole(userRepo, models.UserTypeAdmin)
//   - Multiple roles: RequireRole(userRepo, models.UserTypeAdmin, models.UserTypeModerator)
func RequireRole(userRepo *repository.UserRepository, allowedRoles ...models.UserType) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID from context (set by AuthMiddleware)
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// Fetch user from database to get their role
		var user models.User
		if err := userRepo.DB.Select("id, type").First(&user, "id = ?", userID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Check if user's role is in the allowed roles
		for _, allowedRole := range allowedRoles {
			if user.Type == allowedRole {
				// Store user type in context for handlers to use
				c.Locals("user_type", user.Type)
				return c.Next()
			}
		}

		// User doesn't have required role
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
}
