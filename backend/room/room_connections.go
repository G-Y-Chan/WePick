package room

import (
	"sync"
	"github.com/gorilla/websocket"
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
