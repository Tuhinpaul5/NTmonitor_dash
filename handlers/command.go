package handlers

import (
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

// CommandHandler handles device command requests
type CommandHandler struct {
	registry     *ConnectionRegistry
	commandQueue *CommandQueue
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(registry *ConnectionRegistry) *CommandHandler {
	return &CommandHandler{
		registry:     registry,
		commandQueue: NewCommandQueue(),
	}
}

// CommandRequest represents a command to be sent to a device
type CommandRequest struct {
	Command    string                 `json:"command" validate:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timeout    int                    `json:"timeout,omitempty"` // in seconds
}

// CommandResponse represents the queued command
type CommandResponse struct {
	CommandID string     `json:"command_id"`
	DeviceID  string     `json:"device_id"`
	Command   string     `json:"command"`
	Status    string     `json:"status"` // "queued", "sent", "delivered", "failed"
	QueuedAt  time.Time  `json:"queued_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// QueuedCommand represents a command in the queue
type QueuedCommand struct {
	ID         string
	DeviceID   string
	Command    string
	Parameters map[string]interface{}
	QueuedAt   time.Time
	SentAt     *time.Time
	Status     string
	Error      string
	mu         sync.RWMutex
}

// CommandQueue manages pending commands
type CommandQueue struct {
	commands sync.Map // map[commandID]*QueuedCommand
	counter  uint64
	mu       sync.Mutex
}

// NewCommandQueue creates a new command queue
func NewCommandQueue() *CommandQueue {
	return &CommandQueue{}
}

// Add adds a command to the queue
func (q *CommandQueue) Add(deviceID, command string, parameters map[string]interface{}) *QueuedCommand {
	q.mu.Lock()
	q.counter++
	commandID := time.Now().Format("20060102150405") + "-" + deviceID
	q.mu.Unlock()

	cmd := &QueuedCommand{
		ID:         commandID,
		DeviceID:   deviceID,
		Command:    command,
		Parameters: parameters,
		QueuedAt:   time.Now(),
		Status:     "queued",
	}

	q.commands.Store(commandID, cmd)
	return cmd
}

// Get retrieves a command by ID
func (q *CommandQueue) Get(commandID string) (*QueuedCommand, bool) {
	if val, ok := q.commands.Load(commandID); ok {
		return val.(*QueuedCommand), true
	}
	return nil, false
}

// UpdateStatus updates the command status
func (q *CommandQueue) UpdateStatus(commandID, status string, err error) {
	if cmd, ok := q.Get(commandID); ok {
		cmd.mu.Lock()
		defer cmd.mu.Unlock()

		cmd.Status = status
		if err != nil {
			cmd.Error = err.Error()
		}
		if status == "sent" && cmd.SentAt == nil {
			now := time.Now()
			cmd.SentAt = &now
		}
	}
}

// GetDeviceCommands returns all commands for a specific device
func (q *CommandQueue) GetDeviceCommands(deviceID string) []*QueuedCommand {
	var commands []*QueuedCommand
	q.commands.Range(func(key, value interface{}) bool {
		cmd := value.(*QueuedCommand)
		if cmd.DeviceID == deviceID {
			commands = append(commands, cmd)
		}
		return true
	})
	return commands
}

// Remove removes a command from the queue
func (q *CommandQueue) Remove(commandID string) {
	q.commands.Delete(commandID)
}

// SendCommand sends a command to a device
// @Summary Send command to device
// @Description Queue and send a command to a specific device
// @Tags Commands
// @Accept json
// @Produce json
// @Param device_id path string true "Device ID"
// @Param command body CommandRequest true "Command details"
// @Success 200 {object} CommandResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/devices/{device_id}/command [post]
func (h *CommandHandler) SendCommand(c fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "device_id is required",
		})
	}

	var req CommandRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Command == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "command is required",
		})
	}

	// Add command to queue
	queuedCmd := h.commandQueue.Add(deviceID, req.Command, req.Parameters)

	// Try to send immediately if device is connected
	if dc, connected := h.registry.Get(deviceID); connected {
		msg := WebSocketMessage{
			Type:      "command",
			DeviceID:  deviceID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"command_id": queuedCmd.ID,
				"command":    req.Command,
				"parameters": req.Parameters,
			},
		}

		if err := dc.Conn.WriteJSON(msg); err != nil {
			log.Printf("Error sending command to device %s: %v", deviceID, err)
			h.commandQueue.UpdateStatus(queuedCmd.ID, "failed", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to send command to device",
				"details": err.Error(),
			})
		}

		h.commandQueue.UpdateStatus(queuedCmd.ID, "sent", nil)
		log.Printf("Command %s sent to device %s", queuedCmd.ID, deviceID)
	} else {
		log.Printf("Device %s not connected, command %s queued", deviceID, queuedCmd.ID)
	}

	// Prepare response
	response := CommandResponse{
		CommandID: queuedCmd.ID,
		DeviceID:  deviceID,
		Command:   req.Command,
		Status:    queuedCmd.Status,
		QueuedAt:  queuedCmd.QueuedAt,
		SentAt:    queuedCmd.SentAt,
	}

	return c.JSON(response)
}

// GetCommandStatus gets the status of a command
// @Summary Get command status
// @Description Get the status of a previously sent command
// @Tags Commands
// @Produce json
// @Param command_id path string true "Command ID"
// @Success 200 {object} CommandResponse
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/commands/{command_id} [get]
func (h *CommandHandler) GetCommandStatus(c fiber.Ctx) error {
	commandID := c.Params("command_id")
	if commandID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "command_id is required",
		})
	}

	cmd, exists := h.commandQueue.Get(commandID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Command not found",
		})
	}

	cmd.mu.RLock()
	defer cmd.mu.RUnlock()

	response := CommandResponse{
		CommandID: cmd.ID,
		DeviceID:  cmd.DeviceID,
		Command:   cmd.Command,
		Status:    cmd.Status,
		QueuedAt:  cmd.QueuedAt,
		SentAt:    cmd.SentAt,
		Error:     cmd.Error,
	}

	return c.JSON(response)
}

// GetDeviceCommands gets all commands for a device
// @Summary Get device commands
// @Description Get all commands (queued, sent, failed) for a specific device
// @Tags Commands
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {array} CommandResponse
// @Failure 400 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/devices/{device_id}/commands [get]
func (h *CommandHandler) GetDeviceCommands(c fiber.Ctx) error {
	deviceID := c.Params("device_id")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "device_id is required",
		})
	}

	commands := h.commandQueue.GetDeviceCommands(deviceID)

	var responses []CommandResponse
	for _, cmd := range commands {
		cmd.mu.RLock()
		responses = append(responses, CommandResponse{
			CommandID: cmd.ID,
			DeviceID:  cmd.DeviceID,
			Command:   cmd.Command,
			Status:    cmd.Status,
			QueuedAt:  cmd.QueuedAt,
			SentAt:    cmd.SentAt,
			Error:     cmd.Error,
		})
		cmd.mu.RUnlock()
	}

	return c.JSON(responses)
}

// RetryCommand retries a failed command
// @Summary Retry command
// @Description Retry a failed or queued command
// @Tags Commands
// @Produce json
// @Param command_id path string true "Command ID"
// @Success 200 {object} CommandResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /api/commands/{command_id}/retry [post]
func (h *CommandHandler) RetryCommand(c fiber.Ctx) error {
	commandID := c.Params("command_id")
	if commandID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "command_id is required",
		})
	}

	cmd, exists := h.commandQueue.Get(commandID)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Command not found",
		})
	}

	// Check if device is connected
	dc, connected := h.registry.Get(cmd.DeviceID)
	if !connected {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Device not connected",
		})
	}

	// Send command
	msg := WebSocketMessage{
		Type:      "command",
		DeviceID:  cmd.DeviceID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"command_id": cmd.ID,
			"command":    cmd.Command,
			"parameters": cmd.Parameters,
		},
	}

	if err := dc.Conn.WriteJSON(msg); err != nil {
		h.commandQueue.UpdateStatus(cmd.ID, "failed", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to send command to device",
			"details": err.Error(),
		})
	}

	h.commandQueue.UpdateStatus(cmd.ID, "sent", nil)

	cmd.mu.RLock()
	defer cmd.mu.RUnlock()

	response := CommandResponse{
		CommandID: cmd.ID,
		DeviceID:  cmd.DeviceID,
		Command:   cmd.Command,
		Status:    cmd.Status,
		QueuedAt:  cmd.QueuedAt,
		SentAt:    cmd.SentAt,
	}

	return c.JSON(response)
}
