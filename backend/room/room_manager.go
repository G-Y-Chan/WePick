package room

import (
	"fmt"
	"math/rand"
	"sync"
	"github.com/gorilla/websocket"
	"backend/util"
)

type RoomManager struct {
	mu        sync.RWMutex
	rooms     map[string]*Room
	max       int
}

func NewRoomManager(max int) *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
		max: max,
	}
}

func (rm *RoomManager) GenerateRoomCodeLocked() string {
	// need to handle the case where all possible codes are taken
	var code string

	for true {
		code = fmt.Sprintf("%06d", rand.Intn(rm.max))
		_, exists := rm.rooms[code]
		if !exists {
			break
		}
	}

	return code
}

func (rm *RoomManager) CreateRoom(code string) *Room {
	return &Room{
		Code: code,
		Started: false,
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (rm *RoomManager) AddRoom() string {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	code := rm.GenerateRoomCodeLocked()
	room := rm.CreateRoom(code)
	rm.rooms[code] = room
	return code
}

func (rm *RoomManager) StartRoom(code string) (bool, error) {
	rm.mu.RLock()
	room, exists := rm.rooms[code]
	rm.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("invalid room code")
	}

	room.mu.Lock()
	if room.Started {
		room.mu.Unlock()
		return false, nil
	}
	room.Started = true

	clients := make([]*websocket.Conn, 0, len(room.clients))
	for c := range room.clients {
		clients = append(clients, c)
	}
	room.mu.Unlock()

	msg := util.Message{Header: "START"}
	for _, c := range clients {
		_ = c.WriteJSON(msg)
	}

	return true, nil
}

func (rm *RoomManager) ValidateRoomJoin(code string) (bool, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.mu.RLock()
	room, exists := rm.rooms[code]
	rm.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("invalid room code")
	}

	if room.Started {
		return false, nil
	}

	return true, nil
}

func (rm *RoomManager) RegisterConn(roomCode string, conn *websocket.Conn) (err error) {
	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	room.clients[conn] = struct{}{}
	room.mu.Unlock()

	return nil
}

func (rm *RoomManager) UnregisterConn(roomCode string, conn *websocket.Conn) {
	rm.mu.RLock()
	room, exists := rm.rooms[roomCode]
	rm.mu.RUnlock()
	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.clients, conn)
	room.mu.Unlock()
}
