# WebSocket Gateway Documentation

## Overview

The NTMonitor WebSocket Gateway provides real-time bidirectional communication between IoT devices and the server. It includes connection management, health monitoring, and command dispatch capabilities.

## Architecture Components

### 1. Connection Registry (`handlers/registry.go`)
- **Purpose**: Manages all active WebSocket connections
- **Features**:
  - Thread-safe connection storage using `sync.Map`
  - Last seen timestamp tracking
  - Missed ping counter
  - Connection health monitoring
  - Automatic cleanup of stale connections

### 2. WebSocket Gateway Handler (`handlers/gateway_ws.go`)
- **Purpose**: Handles WebSocket connection lifecycle
- **Features**:
  - JWT-based authentication
  - Device registration/unregistration
  - Message routing (ping/pong, data, status, command)
  - Connection statistics
  - Health monitoring with automatic ping/pong

### 3. Command Handler (`handlers/command.go`)
- **Purpose**: Manages device command queue and dispatch
- **Features**:
  - Command queuing for offline devices
  - Immediate dispatch for online devices
  - Command status tracking
  - Retry mechanism for failed commands

## API Endpoints

### WebSocket Connection

**Endpoint**: `GET /api/gateway/ws`

**Authentication**: JWT token required (Bearer token in header or `token` query parameter)

**Query Parameters**:
- `device_id` (optional): Device identifier. If not provided, extracted from JWT claims.
- `token` (optional): JWT token for authentication

**Example Connection**:
```javascript
// Browser WebSocket client
const token = "your-jwt-token";
const deviceId = "device-123";
const ws = new WebSocket(`ws://localhost:8000/api/gateway/ws?device_id=${deviceId}&token=${token}`);

ws.onopen = () => {
  console.log('Connected to gateway');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

### Message Format

All messages follow this JSON structure:

```json
{
  "type": "ping|pong|command|data|status",
  "device_id": "device-123",
  "timestamp": "2026-07-12T10:30:00Z",
  "payload": {
    // Message-specific data
  }
}
```

#### Message Types

1. **ping/pong**: Health check messages
2. **command**: Commands sent from server to device
3. **data**: Telemetry/sensor data from device
4. **status**: Status updates

### REST API Endpoints

#### 1. Send Command to Device

**Endpoint**: `POST /api/devices/:device_id/command`

**Authentication**: Session-based (cookie)

**Request Body**:
```json
{
  "command": "restart",
  "parameters": {
    "delay": 5,
    "mode": "soft"
  },
  "timeout": 30
}
```

**Response**:
```json
{
  "command_id": "20260712103000-device-123",
  "device_id": "device-123",
  "command": "restart",
  "status": "sent",
  "queued_at": "2026-07-12T10:30:00Z",
  "sent_at": "2026-07-12T10:30:01Z"
}
```

**Status Values**:
- `queued`: Command is queued (device offline)
- `sent`: Command sent to device
- `delivered`: Device acknowledged receipt
- `failed`: Failed to send

#### 2. Get Command Status

**Endpoint**: `GET /api/commands/:command_id`

**Authentication**: Session-based

**Response**:
```json
{
  "command_id": "20260712103000-device-123",
  "device_id": "device-123",
  "command": "restart",
  "status": "sent",
  "queued_at": "2026-07-12T10:30:00Z",
  "sent_at": "2026-07-12T10:30:01Z",
  "error": ""
}
```

#### 3. Get Device Commands

**Endpoint**: `GET /api/devices/:device_id/commands`

**Authentication**: Session-based

**Response**: Array of command objects

#### 4. Retry Command

**Endpoint**: `POST /api/commands/:command_id/retry`

**Authentication**: Session-based

#### 5. Get Connection Statistics

**Endpoint**: `GET /api/gateway/stats`

**Authentication**: JWT-based

**Response**:
```json
{
  "total_connections": 5,
  "devices": [
    {
      "device_id": "device-123",
      "last_seen": "2026-07-12T10:35:00Z",
      "missed_pings": 0,
      "connected": 300.5
    }
  ]
}
```

#### 6. Disconnect Device

**Endpoint**: `POST /api/gateway/devices/:device_id/disconnect`

**Authentication**: JWT-based

## Health Monitoring

The gateway includes automatic health monitoring with configurable parameters:

```go
// In main.go
gatewayHandler.StartHealthMonitor(
    30*time.Second,  // Ping interval
    60*time.Second,  // Cleanup interval
    3,               // Max missed pings before disconnect
)
```

**Features**:
- Automatic ping every 30 seconds
- Cleanup of stale connections every 60 seconds
- Disconnect after 3 missed pings
- Configurable timeouts

## Device Client Example

### Go Client Example

```go
package main

import (
    "encoding/json"
    "log"
    "github.com/gorilla/websocket"
)

type Message struct {
    Type      string                 `json:"type"`
    DeviceID  string                 `json:"device_id"`
    Timestamp time.Time              `json:"timestamp"`
    Payload   map[string]interface{} `json:"payload,omitempty"`
}

func main() {
    deviceID := "device-123"
    token := "your-jwt-token"
    
    url := fmt.Sprintf("ws://localhost:8000/api/gateway/ws?device_id=%s&token=%s", 
        deviceID, token)
    
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        log.Fatal("Connection error:", err)
    }
    defer conn.Close()
    
    // Handle incoming messages
    go func() {
        for {
            var msg Message
            err := conn.ReadJSON(&msg)
            if err != nil {
                log.Println("Read error:", err)
                return
            }
            
            switch msg.Type {
            case "ping":
                // Respond with pong
                pong := Message{
                    Type:      "pong",
                    DeviceID:  deviceID,
                    Timestamp: time.Now(),
                }
                conn.WriteJSON(pong)
                
            case "command":
                // Handle command
                log.Printf("Received command: %v", msg.Payload)
                // Execute command and send response
                
            case "status":
                log.Printf("Status: %v", msg.Payload)
            }
        }
    }()
    
    // Send periodic data
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        data := Message{
            Type:      "data",
            DeviceID:  deviceID,
            Timestamp: time.Now(),
            Payload: map[string]interface{}{
                "temperature": 25.5,
                "humidity":    60,
                "status":      "online",
            },
        }
        conn.WriteJSON(data)
    }
}
```

### Python Client Example

```python
import asyncio
import websockets
import json
from datetime import datetime

async def device_client(device_id, token):
    url = f"ws://localhost:8000/api/gateway/ws?device_id={device_id}&token={token}"
    
    async with websockets.connect(url) as websocket:
        # Receive welcome message
        welcome = await websocket.recv()
        print(f"Welcome: {welcome}")
        
        # Handle messages
        async def receive():
            async for message in websocket:
                data = json.loads(message)
                
                if data['type'] == 'ping':
                    # Respond with pong
                    pong = {
                        'type': 'pong',
                        'device_id': device_id,
                        'timestamp': datetime.utcnow().isoformat()
                    }
                    await websocket.send(json.dumps(pong))
                
                elif data['type'] == 'command':
                    print(f"Command received: {data['payload']}")
                    # Execute command
        
        # Start receiving
        asyncio.create_task(receive())
        
        # Send periodic data
        while True:
            data = {
                'type': 'data',
                'device_id': device_id,
                'timestamp': datetime.utcnow().isoformat(),
                'payload': {
                    'temperature': 25.5,
                    'humidity': 60,
                    'status': 'online'
                }
            }
            await websocket.send(json.dumps(data))
            await asyncio.sleep(10)

# Run
asyncio.run(device_client("device-123", "your-jwt-token"))
```

## Configuration

### Environment Variables

No additional environment variables are required. The gateway uses existing JWT configuration from your `.env` file.

### Customization

You can customize health monitoring parameters in `main.go`:

```go
gatewayHandler.StartHealthMonitor(
    pingInterval,    // time.Duration
    cleanupInterval, // time.Duration
    maxMissedPings,  // int
)
```

## Security Considerations

1. **Authentication**: All WebSocket connections require valid JWT tokens
2. **Authorization**: Command endpoints use session-based authentication
3. **Connection Limits**: Consider implementing rate limiting
4. **Message Validation**: Validate all incoming messages
5. **TLS**: Use WSS (WebSocket Secure) in production

## Monitoring & Debugging

### Check Active Connections

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8000/api/gateway/stats
```

### View Logs

The gateway logs all connection events:
- Device connections/disconnections
- Message sending/receiving
- Health check failures
- Command dispatch

## Future Enhancements

1. **Message Persistence**: Store missed messages for offline devices
2. **Load Balancing**: Distribute connections across multiple servers
3. **Metrics**: Prometheus metrics for monitoring
4. **Rate Limiting**: Prevent message flooding
5. **Message Acknowledgment**: Confirm command delivery
6. **WebSocket Compression**: Reduce bandwidth usage

## Troubleshooting

### Device Can't Connect

1. Verify JWT token is valid
2. Check `device_id` is provided
3. Ensure firewall allows WebSocket connections
4. Check server logs for connection errors

### Commands Not Received

1. Verify device is connected (check `/api/gateway/stats`)
2. Check command queue status
3. Review device message handling logic
4. Check for network issues

### High Missed Ping Count

1. Check network stability
2. Verify device is responding to pings
3. Increase ping timeout if network is slow
4. Check device processing capacity
