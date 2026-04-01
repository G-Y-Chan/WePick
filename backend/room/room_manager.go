package room

import (
	"fmt"
	"math/rand"
	"sync"
	"github.com/gorilla/websocket"
)

type RoomManager struct {
	mu        		sync.RWMutex
	roomRepository 	*RoomRepository
	rooms 			map[string]*RoomConnections
	max       		int
}

func NewRoomManager(max int) *RoomManager {
	return &RoomManager{
		roomRepository: NewRoomRepository(),
		rooms: make(map[string]*RoomConnections),
		max: max,
	}
}

func (rm *RoomManager) GenerateRoomCode() string {
	return fmt.Sprintf("%06d", rand.Intn(rm.max))
}

func (rm *RoomManager) AddRoom() (string, error) {
	for {
		code := rm.GenerateRoomCode()

		roomCreated, err := rm.roomRepository.CreateRoom(code)
		if err != nil {
			return "", err
		}

		if roomCreated {
			return code, nil
		}

		// else: collision → try again
	}
}

func (rm *RoomManager) StartRoom(code string) (bool, error) {
	// rm.mu.RLock()
	// room, exists := rm.rooms[code]
	// rm.mu.RUnlock()
	// if !exists {
	// 	return false, fmt.Errorf("invalid room code")
	// }

	// room.mu.Lock()
	// if room.Started {
	// 	room.mu.Unlock()
	// 	return false, nil
	// }
	// room.Started = true

	// clients := make([]*websocket.Conn, 0, len(room.clients))
	// for c := range room.clients {
	// 	clients = append(clients, c)
	// }
	// room.mu.Unlock()

	// msg := util.Message{Header: "START"}
	// for _, c := range clients {
	// 	_ = c.WriteJSON(msg)
	// }

	return true, nil
}

func (rm *RoomManager) ValidateRoomJoin(code string) (bool, error) {
	started, err := rm.roomRepository.CheckIfRoomStarted(code)
	if err != nil {
		return false, err
	}
	if started {
		return false, nil
	}

	return true, nil
}

func (rm *RoomManager) RegisterConn(code string, conn *websocket.Conn) error {
	exists, err := rm.roomRepository.CheckIfRoomExists(code)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("room not found")
	}

	rm.mu.Lock()
	roomConnections, exists := rm.rooms[code]
	if !exists {
		roomConnections = NewRoomConnections()
		rm.rooms[code] = roomConnections
	}
	rm.mu.Unlock()

	roomConnections.Add(conn)
	return nil
}

func (rm *RoomManager) UnregisterConn(code string, conn *websocket.Conn) {
	rm.mu.RLock()
	roomConnections, exists := rm.rooms[code]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	roomConnections.Remove(conn)
}
