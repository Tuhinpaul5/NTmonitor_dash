package router

import (
	"NTMonitor/handlers"

	"github.com/gofiber/fiber/v3"
)

// RegisterAuthRoutes handles all /api/auth/... endpoints
func NodeRoutes(api fiber.Router, h *handlers.NodeHandler, jwtMiddleware fiber.Handler) {
	auth := api.Group("/node")

	auth.Get("/nodes", h.GetAllNodes)
	auth.Get("/nodes/:id", h.GetNodeByID)
	auth.Post("/nodes", h.AddNode)
}
