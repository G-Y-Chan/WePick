package room

import (
	"sync"
	"github.com/gorilla/websocket"
	"backend/util"
	"fmt"
)

type RoomConnections struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]struct{}
}

func NewRoomConnections() *RoomConnections {
	return &RoomConnections{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (rc *RoomConnections) Add(conn *websocket.Conn) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.clients[conn] = struct{}{}
}

func (rc *RoomConnections) Remove(conn *websocket.Conn) {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    delete(rc.clients, conn)
}

func (rc *RoomConnections) BroadcastRoomStarted() {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	fmt.Println("Broadcasting room started event to", len(rc.clients), "clients")
	msg := util.Message{Header: "START"}
	for conn := range rc.clients {
		conn.WriteJSON(msg)
	}
}
