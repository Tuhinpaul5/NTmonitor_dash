package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// GatewayWSHandler handles WebSocket connections for IoT devices
type GatewayWSHandler struct {
	registry *ConnectionRegistry
}

// NewGatewayWSHandler creates a new WebSocket gateway handler
func NewGatewayWSHandler(registry *ConnectionRegistry) *GatewayWSHandler {
	return &GatewayWSHandler{
		registry: registry,
	}
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type      string                 `json:"type"`      // "ping", "pong", "command", "data", "status"
	DeviceID  string                 `json:"device_id"` // Device identifier
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// UpgradeCheck validates WebSocket upgrade requests
// This should be used before upgrading to WebSocket
func (h *GatewayWSHandler) UpgradeCheck() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check if it's a WebSocket upgrade request
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		// Optionally validate JWT token from query parameter or header
		// The JWT middleware will handle this if you apply it before this handler
		return c.Next()
	}
}

// HandleWebSocket handles the WebSocket connection lifecycle
func (h *GatewayWSHandler) HandleWebSocket(c *websocket.Conn) {
	// Extract device ID from query parameter or JWT claims
	deviceID := c.Query("device_id")

	// If device_id not in query, try to get from JWT token stored in locals
	if deviceID == "" {
		// Get token from Fiber context (set by JWT middleware)
		if token := c.Locals("user"); token != nil {
			if jwtToken, ok := token.(*jwt.Token); ok {
				if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
					if id, exists := claims["device_id"]; exists {
						deviceID = id.(string)
					} else if id, exists := claims["sub"]; exists {
						deviceID = id.(string)
					}
				}
			}
		}
	}

	if deviceID == "" {
		log.Println("WebSocket connection rejected: missing device_id")
		c.WriteMessage(websocket.TextMessage, []byte(`{"error":"missing device_id"}`))
		c.Close()
		return
	}

	// Register connection
	dc := h.registry.Register(deviceID, c)
	log.Printf("Device %s connected. Total connections: %d", deviceID, h.registry.Count())

	// Ensure cleanup on disconnect
	defer func() {
		h.registry.Unregister(deviceID)
		log.Printf("Device %s disconnected. Total connections: %d", deviceID, h.registry.Count())
	}()

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type:      "status",
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"status":  "connected",
			"message": "Successfully connected to NTMonitor gateway",
		},
	}
	if err := c.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Error sending welcome message to %s: %v", deviceID, err)
		return
	}

	// Message handling loop
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for device %s: %v", deviceID, err)
			}
			break
		}

		// Update last seen timestamp
		dc.UpdateLastSeen()

		// Handle different message types
		if messageType == websocket.TextMessage {
			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error parsing message from device %s: %v", deviceID, err)
				continue
			}

			// Handle message based on type
			switch msg.Type {
			case "ping":
				// Respond with pong
				pongMsg := WebSocketMessage{
					Type:      "pong",
					DeviceID:  deviceID,
					Timestamp: time.Now(),
				}
				if err := c.WriteJSON(pongMsg); err != nil {
					log.Printf("Error sending pong to device %s: %v", deviceID, err)
					return
				}

			case "pong":
				// Device responded to our ping
				log.Printf("Received pong from device %s", deviceID)

			case "data":
				// Handle device data (telemetry, sensor readings, etc.)
				log.Printf("Received data from device %s: %+v", deviceID, msg.Payload)
				// TODO: Process and store device data in database

			case "status":
				// Handle device status updates
				log.Printf("Device %s status: %+v", deviceID, msg.Payload)

			default:
				log.Printf("Unknown message type '%s' from device %s", msg.Type, deviceID)
			}
		} else if messageType == websocket.BinaryMessage {
			log.Printf("Received binary message from device %s (%d bytes)", deviceID, len(message))
			// TODO: Handle binary data if needed
		}
	}
}

// GetConnectionStats returns connection statistics
// @Summary Get connection statistics
// @Description Get statistics about active WebSocket connections
// @Tags Gateway
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/gateway/stats [get]
func (h *GatewayWSHandler) GetConnectionStats(c fiber.Ctx) error {
	stats := h.registry.GetConnectionStats()
	return c.JSON(stats)
}

// DisconnectDevice forcefully disconnects a device
// @Summary Disconnect a device
// @Description Forcefully disconnect a device from the gateway
// @Tags Gateway
// @Param device_id path string true "Device ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/gateway/devices/{device_id}/disconnect [post]
func (h *GatewayWSHandler) DisconnectDevice(c fiber.Ctx) error {
	deviceID := c.Params("device_id")

	dc, exists := h.registry.Get(deviceID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Device not connected",
		})
	}

	// Send disconnect message before closing
	disconnectMsg := WebSocketMessage{
		Type:      "status",
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"status":  "disconnected",
			"message": "Connection closed by server",
		},
	}
	dc.Conn.WriteJSON(disconnectMsg)

	// Close connection
	dc.Conn.Close()
	h.registry.Unregister(deviceID)

	return c.JSON(fiber.Map{
		"message":   "Device disconnected successfully",
		"device_id": deviceID,
	})
}

// StartHealthMonitor starts a background goroutine to monitor connection health
func (h *GatewayWSHandler) StartHealthMonitor(pingInterval time.Duration, cleanupInterval time.Duration, maxMissedPings int) {
	// Ping routine - sends ping to all connected devices
	go func() {
		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for range ticker.C {
			connections := h.registry.GetAll()
			for _, dc := range connections {
				pingMsg := WebSocketMessage{
					Type:      "ping",
					DeviceID:  dc.DeviceID,
					Timestamp: time.Now(),
				}

				if err := dc.Conn.WriteJSON(pingMsg); err != nil {
					log.Printf("Error sending ping to device %s: %v", dc.DeviceID, err)
					dc.IncrementMissedPings()
				}
			}
		}
	}()

	// Cleanup routine - removes stale connections
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			timeout := pingInterval * time.Duration(maxMissedPings)
			removed := h.registry.CleanupStaleConnections(timeout, maxMissedPings)
			if len(removed) > 0 {
				log.Printf("Cleaned up %d stale connections: %v", len(removed), removed)
			}
		}
	}()

	log.Printf("Health monitor started (ping: %v, cleanup: %v, max missed pings: %d)",
		pingInterval, cleanupInterval, maxMissedPings)
}
