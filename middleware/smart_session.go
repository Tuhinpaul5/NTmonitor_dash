package middleware

import (
	"NTMonitor/repository"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/postgres/v3"
)

// SmartSessionMiddleware only creates sessions when needed
func SmartSessionMiddleware(cfg *fiber.Config, dbURL string) fiber.Handler {
	// Configure session storage
	storage := postgres.New(postgres.Config{
		ConnectionURI: dbURL,
		Table:         "fiber_sessions", // Different table from your custom sessions
		GCInterval:    10 * time.Minute,
	})

	sessionStore := session.New(session.Config{
		Storage:         storage,
		CookieSecure:    true,
		CookieHTTPOnly:  true,
		CookieSameSite:  "Lax",
		IdleTimeout:     30 * time.Minute,
		AbsoluteTimeout: 24 * time.Hour,
	})

	return func(c fiber.Ctx) error {
		// Only create sessions for specific routes or conditions
		path := c.Path()

		// Skip session creation for static files, docs, health checks
		if shouldSkipSession(path) {
			return c.Next()
		}

		// Apply session middleware
		return sessionStore(c)
	}
}

func shouldSkipSession(path string) bool {
	skipPaths := []string{
		"/health",
		"/docs",
		"/swagger",
		"/test-swagger",
	}

	for _, skipPath := range skipPaths {
		if len(path) >= len(skipPath) && path[:len(skipPath)] == skipPath {
			return true
		}
	}
	return false
}

// ConditionalSessionMiddleware creates sessions only after successful auth
func ConditionalSessionMiddleware(sessionRepo *repository.SessionRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check if user is already authenticated via your custom sessions
		sessionToken := c.Cookies("session_token")
		if sessionToken != "" {
			session, err := sessionRepo.FindByToken(sessionToken)
			if err == nil {
				// User is authenticated, store user ID
				c.Locals("user_id", session.UserID)
				c.Locals("authenticated", true)
				return c.Next()
			}
		}

		// Check if this is a login attempt
		if c.Path() == "/api/auth/login" && c.Method() == "POST" {
			// Let the login handler decide whether to create a session
			return c.Next()
		}

		// For protected routes, require authentication
		if isProtectedRoute(c.Path()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		return c.Next()
	}
}

func isProtectedRoute(path string) bool {
	protectedPaths := []string{
		"/api/users",
		"/api/profile",
		"/api/admin",
	}

	for _, protectedPath := range protectedPaths {
		if len(path) >= len(protectedPath) && path[:len(protectedPath)] == protectedPath {
			return true
		}
	}
	return false
}
