package wshub

import (
	"errors"
	"log/slog"
	"sync"

	"backend/internal/domain/room"
)

var (
	errEmptyRoomCode = errors.New("room code must not be empty")
	errNilClient     = errors.New("client must not be nil")
)

// Hub implements room.Broadcaster and fans realtime events out to the clients
// currently connected to a room. It knows only about the room.Client and
// room.OutboundEvent abstractions; it has no knowledge of WebSockets beyond
// what Conn encapsulates.
type Hub struct {
	mu     sync.RWMutex
	rooms  map[room.Code]map[room.Client]struct{}
	logger *slog.Logger
}

// NewHub constructs a Hub. A nil logger falls back to slog.Default.
func NewHub(logger *slog.Logger) *Hub {
	if logger == nil {
		logger = slog.Default()
	}

	return &Hub{
		rooms:  make(map[room.Code]map[room.Client]struct{}),
		logger: logger,
	}
}

// Register adds client to the room's connection set.
func (h *Hub) Register(code room.Code, client room.Client) error {
	if code == "" {
		return errEmptyRoomCode
	}
	if client == nil {
		return errNilClient
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	clients, exists := h.rooms[code]
	if !exists {
		clients = make(map[room.Client]struct{})
		h.rooms[code] = clients
	}
	clients[client] = struct{}{}

	return nil
}

// Unregister removes client from the room. If the room's client set becomes
// empty, the room entry is deleted from the map, closing the unbounded memory
// growth gap identified in the architectural audit.
func (h *Hub) Unregister(code room.Code, client room.Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	clients, exists := h.rooms[code]
	if !exists {
		return
	}

	delete(clients, client)
	if len(clients) == 0 {
		delete(h.rooms, code)
	}
}

// BroadcastRoomStarted sends a START event to every client in the room.
func (h *Hub) BroadcastRoomStarted(code room.Code) {
	h.broadcast(code, room.OutboundEvent{Type: room.OutboundStart})
}

// BroadcastMajorityFound sends a MAJORITY_FOUND event for cardID to every
// client in the room.
func (h *Hub) BroadcastMajorityFound(code room.Code, cardID string) {
	h.broadcast(code, room.OutboundEvent{
		Type:   room.OutboundMajorityFound,
		CardID: cardID,
	})
}

// broadcast snapshots the room's clients under a read lock, then delivers the
// event without holding the lock so a slow Send cannot block other rooms'
// fan-out. A client whose Send returns an error (full buffer or closed
// connection) is evicted.
func (h *Hub) broadcast(code room.Code, evt room.OutboundEvent) {
	clients := h.snapshot(code)

	for _, client := range clients {
		if err := client.Send(evt); err != nil {
			h.logger.Warn(
				"dropping slow websocket client",
				"room_code", string(code),
				"conn_id", client.ID(),
				"err", err,
			)
			h.Unregister(code, client)
		}
	}
}

// snapshot returns a copy of the room's current client set.
func (h *Hub) snapshot(code room.Code) []room.Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.rooms[code]
	snapshot := make([]room.Client, 0, len(clients))
	for client := range clients {
		snapshot = append(snapshot, client)
	}
	return snapshot
}

// Compile-time contract check.
var _ room.Broadcaster = (*Hub)(nil)
