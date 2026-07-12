package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

// RegisterGatewayRoutes registers all gateway-related routes
func RegisterGatewayRoutes(
	api fiber.Router,
	gatewayHandler *handlers.GatewayWSHandler,
	commandHandler *handlers.CommandHandler,
	jwtMiddleware fiber.Handler,
) {
	gateway := api.Group("/gateway")

	// WebSocket endpoint with JWT authentication
	// Clients can pass JWT token via query parameter: ?token=xxx
	// or via Authorization header (validated by middleware)
	gateway.Use("/ws", jwtMiddleware, gatewayHandler.UpgradeCheck())
	gateway.Get("/ws", websocket.New(gatewayHandler.HandleWebSocket))

	// Gateway management endpoints (protected)
	gateway.Get("/stats", jwtMiddleware, gatewayHandler.GetConnectionStats)
	gateway.Post("/devices/:device_id/disconnect", jwtMiddleware, gatewayHandler.DisconnectDevice)
}

// RegisterCommandRoutes registers all command-related routes
func RegisterCommandRoutes(
	api fiber.Router,
	commandHandler *handlers.CommandHandler,
	authMiddleware fiber.Handler,
) {
	// Device command endpoints (protected)
	api.Post("/devices/:device_id/command", authMiddleware, commandHandler.SendCommand)
	api.Get("/devices/:device_id/commands", authMiddleware, commandHandler.GetDeviceCommands)

	// Command management endpoints (protected)
	api.Get("/commands/:command_id", authMiddleware, commandHandler.GetCommandStatus)
	api.Post("/commands/:command_id/retry", authMiddleware, commandHandler.RetryCommand)
}
