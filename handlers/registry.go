package handlers

import (
	"sync"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
)

// DeviceConnection represents a connected device with health tracking
type DeviceConnection struct {
	Conn        *websocket.Conn
	DeviceID    string
	LastSeen    time.Time
	MissedPings int
	mu          sync.RWMutex
}

// UpdateLastSeen updates the last seen timestamp
func (dc *DeviceConnection) UpdateLastSeen() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.LastSeen = time.Now()
	dc.MissedPings = 0
}

// IncrementMissedPings increments the missed ping counter
func (dc *DeviceConnection) IncrementMissedPings() int {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.MissedPings++
	return dc.MissedPings
}

// GetMissedPings returns the current missed ping count
func (dc *DeviceConnection) GetMissedPings() int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.MissedPings
}

// ConnectionRegistry manages all active WebSocket connections
type ConnectionRegistry struct {
	connections sync.Map // map[string]*DeviceConnection
	mu          sync.RWMutex
}

// NewConnectionRegistry creates a new connection registry
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{}
}

// Register adds a new connection to the registry
func (r *ConnectionRegistry) Register(deviceID string, conn *websocket.Conn) *DeviceConnection {
	dc := &DeviceConnection{
		Conn:        conn,
		DeviceID:    deviceID,
		LastSeen:    time.Now(),
		MissedPings: 0,
	}
	r.connections.Store(deviceID, dc)
	return dc
}

// Unregister removes a connection from the registry
func (r *ConnectionRegistry) Unregister(deviceID string) {
	r.connections.Delete(deviceID)
}

// Get retrieves a connection by device ID
func (r *ConnectionRegistry) Get(deviceID string) (*DeviceConnection, bool) {
	if val, ok := r.connections.Load(deviceID); ok {
		return val.(*DeviceConnection), true
	}
	return nil, false
}

// GetAll returns all active connections
func (r *ConnectionRegistry) GetAll() []*DeviceConnection {
	var connections []*DeviceConnection
	r.connections.Range(func(key, value interface{}) bool {
		connections = append(connections, value.(*DeviceConnection))
		return true
	})
	return connections
}

// Count returns the number of active connections
func (r *ConnectionRegistry) Count() int {
	count := 0
	r.connections.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// IsConnected checks if a device is currently connected
func (r *ConnectionRegistry) IsConnected(deviceID string) bool {
	_, ok := r.connections.Load(deviceID)
	return ok
}

// CleanupStaleConnections removes connections that haven't been seen for a while
func (r *ConnectionRegistry) CleanupStaleConnections(timeout time.Duration, maxMissedPings int) []string {
	var removed []string
	now := time.Now()

	r.connections.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		dc := value.(*DeviceConnection)

		dc.mu.RLock()
		timeSinceLastSeen := now.Sub(dc.LastSeen)
		missedPings := dc.MissedPings
		dc.mu.RUnlock()

		if timeSinceLastSeen > timeout || missedPings > maxMissedPings {
			dc.Conn.Close()
			r.Unregister(deviceID)
			removed = append(removed, deviceID)
		}

		return true
	})

	return removed
}

// GetConnectionStats returns statistics about connections
func (r *ConnectionRegistry) GetConnectionStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_connections": r.Count(),
		"devices":           []map[string]interface{}{},
	}

	r.connections.Range(func(key, value interface{}) bool {
		deviceID := key.(string)
		dc := value.(*DeviceConnection)

		dc.mu.RLock()
		deviceStats := map[string]interface{}{
			"device_id":    deviceID,
			"last_seen":    dc.LastSeen,
			"missed_pings": dc.MissedPings,
			"connected":    time.Since(dc.LastSeen).Seconds(),
		}
		dc.mu.RUnlock()

		stats["devices"] = append(stats["devices"].([]map[string]interface{}), deviceStats)
		return true
	})

	return stats
}
