package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pingPeriod     = 30 * time.Second
	pongWait       = 60 * time.Second
	closeGracePeriod = 10 * time.Second
)

// Client represents a single WebSocket connection with a server-generated ID.
type Client struct {
	ID       string
	conn     *websocket.Conn
	send     chan []byte
	hub      *Hub
	roomCode string
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(writeWait))
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				slog.Warn("write to client failed", "clientID", c.ID, "room", c.roomCode, "error", err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Hub manages all WebSocket connections per room using a fan-out pattern.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[string]*Client // roomCode -> clientID -> Client
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[string]*Client),
	}
}

func (h *Hub) AddClient(roomCode string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomCode]; !ok {
		h.rooms[roomCode] = make(map[string]*Client)
	}
	h.rooms[roomCode][client.ID] = client
	slog.Info("client registered", "clientID", client.ID, "room", roomCode, "total", len(h.rooms[roomCode]))
}

func (h *Hub) RemoveClient(roomCode, clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.rooms[roomCode]
	if !ok {
		return
	}

	delete(clients, clientID)
	slog.Info("client unregistered", "clientID", clientID, "room", roomCode, "remaining", len(clients))

	// Clean up empty rooms
	if len(clients) == 0 {
		delete(h.rooms, roomCode)
		slog.Info("room connections cleaned up", "room", roomCode)
	}
}

// RegisterConn implements room.RoomHub.
func (h *Hub) RegisterConn(code, clientID string) error {
	// The client is already added via AddClient during WS upgrade.
	// This is a no-op for the hub since the hub doesn't track Redis state.
	return nil
}

// UnregisterConn implements room.RoomHub.
func (h *Hub) UnregisterConn(code, clientID string) {
	h.RemoveClient(code, clientID)
}

// BroadcastRoomStarted implements room.RoomHub.
func (h *Hub) BroadcastRoomStarted(code string) {
	msg, _ := json.Marshal(map[string]string{"Header": "START"})
	h.broadcastToRoom(code, msg)
}

// BroadcastMajorityFound implements room.RoomHub.
func (h *Hub) BroadcastMajorityFound(code, voteID string) {
	msg, _ := json.Marshal(map[string]interface{}{
		"Header": "MAJORITY_FOUND",
		"VoteObj": map[string]string{
			"Id": voteID,
		},
	})
	h.broadcastToRoom(code, msg)
}

// HasRoom implements room.RoomHub.
func (h *Hub) HasRoom(code string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.rooms[code]
	return ok
}

func (h *Hub) broadcastToRoom(code string, message []byte) {
	h.mu.RLock()
	clients := h.rooms[code]
	h.mu.RUnlock()

	if clients == nil {
		return
	}

	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			// Client's send buffer is full; drop them
			slog.Warn("dropping slow client", "clientID", client.ID, "room", code)
			h.RemoveClient(code, client.ID)
			_ = client.conn.Close()
		}
	}
}

// BroadcastStopped notifies the room about a stopped event.
func (h *Hub) BroadcastStopped(code string) {
	msg, _ := json.Marshal(map[string]string{"Header": "STOPPED"})
	h.broadcastToRoom(code, msg)
}

// CloseRoom closes all connections in a room and removes it.
func (h *Hub) CloseRoom(code string) {
	h.mu.Lock()
	clients := h.rooms[code]
	delete(h.rooms, code)
	h.mu.Unlock()

	if clients == nil {
		return
	}

	for _, client := range clients {
		close(client.send)
		_ = client.conn.Close()
	}
	slog.Info("room closed", "room", code, "connections", len(clients))
}