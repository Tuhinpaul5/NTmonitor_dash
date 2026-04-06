package router

import (
	"NTMonitor/config"
	_ "NTMonitor/docs" // <--- CRITICAL: Import your generated docs
	"NTMonitor/handlers"
	"NTMonitor/middleware"
	"NTMonitor/repository"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/gofiber/fiber/v3/middleware/cors" // Add this
)

func SetupRoutes(
	app *fiber.App,
	userHand *handlers.UserHandler,
	authHand *handlers.AuthHandler,
	sessionRepo *repository.SessionRepository,
) {

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000", "http://127.0.0.1:8000"}, // Allow your frontend origin
		AllowHeaders:     []string{"*"}, // Allow all headers
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"*"}, // Expose all headers
	}))

	app.Use(logger.New())

	// 1. Serve the generated Swagger JSON from the /docs folder
	// This makes http://localhost:8000/swagger/swagger.json accessible
	app.Get("/swagger/swagger.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.SendFile("./docs/swagger.json")
	})

	app.Get("/swagger/*", func(c fiber.Ctx) error {
		// Extract the file path from the URL
		filePath := c.Params("*")
		if filePath == "" {
			filePath = "swagger.json"
		}

		// Serve files from the docs directory
		return c.SendFile("./docs/" + filePath)
	})

	// 2. API Documentation UI (with fallback options)
	app.Get("/docs", func(c fiber.Ctx) error {
		html := `<!doctype html>
<html>
  <head>
    <title>NTMonitor API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <div id="scalar-container"></div>
    <script 
      id="api-reference" 
      data-url="/swagger/swagger.json"
      data-configuration='{"theme":"purple","hideDownloadButton":true}'
    ></script>
    <script>
      // Fallback error handling
      window.addEventListener('error', function(e) {
        console.error('Error loading Scalar:', e);
        document.getElementById('scalar-container').innerHTML = 
          '<h1>API Documentation</h1>' +
          '<p>Swagger JSON available at: <a href="/swagger/swagger.json">/swagger/swagger.json</a></p>' +
          '<p>Alternative Swagger UI: <a href="/swagger-ui">Swagger UI</a></p>';
      });
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// 3. Alternative Swagger UI
	app.Get("/swagger-ui", func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
  <head>
    <title>NTMonitor API - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
      SwaggerUIBundle({
        url: '/swagger/swagger.json',
        dom_id: '#swagger-ui',
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.presets.standalone
        ]
      });
    </script>
  </body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// Public Routes
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test route to verify swagger.json is accessible
	app.Get("/test-swagger", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":    "Swagger JSON should be available at /swagger/swagger.json",
			"docs":       "API documentation available at /docs",
			"swagger-ui": "Alternative Swagger UI available at /swagger-ui",
			"endpoints": fiber.Map{
				"health":            "/health",
				"swagger_json":      "/swagger/swagger.json",
				"docs":              "/docs",
				"swagger_ui":        "/swagger-ui",
				"api_auth_login":    "/api/auth/login",
				"api_auth_logout":   "/api/auth/logout",
				"api_auth_register": "/api/auth/register",
				"api_users":         "/api/users",
			},
		})
	})

	// Debug endpoint to test cookies
	app.Get("/test-cookie", func(c fiber.Ctx) error {
		// Set multiple test cookies with different configurations
		c.Cookie(&fiber.Cookie{
			Name:     "test_cookie",
			Value:    "test_value_123",
			HTTPOnly: false, // Allow JavaScript access for testing
			Secure:   false, // Allow HTTP for testing
			SameSite: "Lax",
			Path:     "/",
		})

		c.Cookie(&fiber.Cookie{
			Name:     "test_cookie_2",
			Value:    "another_test_value",
			HTTPOnly: false,
			Secure:   false,
			SameSite: "None",
			Path:     "/",
		})

		c.Cookie(&fiber.Cookie{
			Name:     "test_cookie_3",
			Value:    "third_test_value",
			HTTPOnly: true,
			Secure:   false,
			SameSite: "Strict",
			Path:     "/",
		})

		// Read existing cookies
		sessionToken := c.Cookies("session_token")
		testCookie := c.Cookies("test_cookie")

		return c.JSON(fiber.Map{
			"message":            "Cookie test endpoint",
			"set_test_cookies":   []string{"test_cookie", "test_cookie_2", "test_cookie_3"},
			"read_session_token": sessionToken,
			"read_test_cookie":   testCookie,
			"request_headers":    c.GetReqHeaders(),
		})
	})

	// Simple cookie test without JavaScript
	app.Get("/simple-test", func(c fiber.Ctx) error {
		// Set a simple cookie
		c.Cookie(&fiber.Cookie{
			Name:     "simple_test",
			Value:    "works",
			HTTPOnly: false,
			Secure:   false,
			SameSite: "Lax",
			Path:     "/",
		})

		html := `<!DOCTYPE html>
<html>
<head>
    <title>Simple Cookie Test</title>
</head>
<body>
    <h1>Simple Cookie Test</h1>
    <p>A cookie named 'simple_test' with value 'works' should be set.</p>
    <p>Check your browser's developer tools → Application → Cookies</p>
    <p>Current cookies in JavaScript: <span id="cookies"></span></p>
    <script>
        document.getElementById('cookies').textContent = document.cookie || 'No cookies found';
    </script>
    <br><br>
    <a href="/test-login-form">Test Login Form</a>
</body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// Simple login form for testing
	app.Get("/test-login-form", func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Login Test Form</title>
</head>
<body>
    <h1>Login Test Form</h1>
    <form action="/api/auth/login" method="POST">
        <div>
            <label>Email:</label><br>
            <input type="email" name="email" value="test@example.com" required>
        </div>
        <br>
        <div>
            <label>Password:</label><br>
            <input type="password" name="password" value="password123" required>
        </div>
        <br>
        <button type="submit">Login</button>
    </form>
    <br>
    <p>Current cookies: <span id="cookies"></span></p>
    <script>
        document.getElementById('cookies').textContent = document.cookie || 'No cookies found';
    </script>
</body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})
	app.Get("/test-page", func(c fiber.Ctx) error {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Cookie Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        div { margin: 20px 0; padding: 15px; border: 1px solid #ccc; }
        button { padding: 10px 15px; margin: 5px; }
        input { padding: 8px; margin: 5px; width: 200px; }
        pre { background: #f5f5f5; padding: 10px; overflow-x: auto; }
        .error { color: red; }
        .success { color: green; }
    </style>
</head>
<body>
    <h1>Cookie Test Page</h1>
    
    <div>
        <h2>Test Basic Cookie</h2>
        <button onclick="testBasicCookie()">Test Basic Cookie</button>
        <div id="basic-result"></div>
    </div>

    <div>
        <h2>Test Login</h2>
        <input type="email" id="email" placeholder="Email" value="user@example.com">
        <input type="password" id="password" placeholder="Password" value="password123">
        <button onclick="testLogin()">Login</button>
        <div id="login-result"></div>
    </div>

    <div>
        <h2>Test Protected Route</h2>
        <button onclick="testProtectedRoute()">Test /api/users</button>
        <div id="protected-result"></div>
    </div>

    <div>
        <h2>Current Cookies</h2>
        <button onclick="showCookies()">Show Cookies</button>
        <div id="cookies-result"></div>
    </div>

    <script>
        async function testBasicCookie() {
            try {
                const response = await fetch('/test-cookie', {
                    credentials: 'include'
                });
                const data = await response.json();
                document.getElementById('basic-result').innerHTML = 
                    '<div class="success"><pre>' + JSON.stringify(data, null, 2) + '</pre></div>';
            } catch (error) {
                document.getElementById('basic-result').innerHTML = 
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        async function testLogin() {
            const email = document.getElementById('email').value;
            const password = document.getElementById('password').value;

            try {
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    credentials: 'include',
                    body: JSON.stringify({ email, password })
                });

                const data = await response.json();
                const resultClass = response.ok ? 'success' : 'error';
                document.getElementById('login-result').innerHTML = 
                    '<div class="' + resultClass + '"><pre>Status: ' + response.status + '\\n' + JSON.stringify(data, null, 2) + '</pre></div>';
                
                console.log('Response headers:', [...response.headers.entries()]);
                setTimeout(showCookies, 500);
            } catch (error) {
                document.getElementById('login-result').innerHTML = 
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        async function testProtectedRoute() {
            try {
                const response = await fetch('/api/users', {
                    credentials: 'include'
                });
                const data = await response.json();
                const resultClass = response.ok ? 'success' : 'error';
                document.getElementById('protected-result').innerHTML = 
                    '<div class="' + resultClass + '"><pre>Status: ' + response.status + '\\n' + JSON.stringify(data, null, 2) + '</pre></div>';
            } catch (error) {
                document.getElementById('protected-result').innerHTML = 
                    '<div class="error">Error: ' + error.message + '</div>';
            }
        }

        function showCookies() {
            const cookies = document.cookie || 'No cookies found';
            document.getElementById('cookies-result').innerHTML = 
                '<div><pre>document.cookie: ' + cookies + '</pre></div>';
        }

        window.onload = function() {
            showCookies();
        };
    </script>
</body>
</html>`

		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(html)
	})

	// Create test user endpoint (for development only)
	app.Post("/create-test-user", func(c fiber.Ctx) error {
		// This should only work in development
		if cfg := config.LoadConfig(); cfg.APP_ENV != "local" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "This endpoint only works in local development",
			})
		}

		// You'll need to implement this based on your user creation logic
		return c.JSON(fiber.Map{
			"message": "Test user creation endpoint - implement based on your needs",
			"note":    "You may need to create a user manually in your database for testing",
		})
	})

	// Versioned Group
	api := app.Group("/api")

	// Public routes (no authentication required)
	RegisterAuthRoutes(api, authHand)
	RegisterUserRoutes(api, userHand) // Public user routes if any

	// Protected routes (require authentication)
	authMiddleware := middleware.AuthMiddleware(sessionRepo)
	RegisterProtectedUserRoutes(api, userHand, authMiddleware)
}
