package room

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"github.com/gorilla/websocket"
	"backend/util"
)

type Room struct {
    Code      int
    Started   bool

	clients   map[*websocket.Conn]struct{}
	mu        sync.RWMutex
}

type Service struct {
	mu        sync.RWMutex
	rooms     map[int]*Room
	max       int
}

// NewService creates a new RoomService with a given max code value (exclusive).
func NewService(max int) *Service {
	return &Service{
		rooms: make(map[int]*Room),
		max: max,
	}
}

func (s *Service) GenerateCode() string {
	paddedLength := 6

	s.mu.Lock()
	defer s.mu.Unlock()

	var code int
	code = s.generateRoomCodeLocked()
	s.insertRoomCodeLocked(code)
	if code >= 100000 {
		return strconv.Itoa(code)
	} else {
		var temp = strconv.Itoa(code)
		for len(temp) < paddedLength {
			temp = "0" + temp
		}
		return temp
	}
}

func (s *Service) generateRoomCodeLocked() int {
	var code int
	for true {
		code = rand.Intn(s.max)
		_, exists := s.rooms[code]
		if !exists {
			break
		}
	}
	s.max = s.max - 1
	return code
}

// Logic behind generating new room code guarantees validity of code.
// Thus, no checks required before inserting new code into intervals.
func (s *Service) insertRoomCodeLocked(code int) {
	room := &Room{
        Code:    code,
        Started: false,
		clients: make(map[*websocket.Conn]struct{}),
    }
	s.rooms[code] = room
}

func (s *Service) VerifyCode(roomCode string) bool {
	var intCode, err = strconv.Atoi(roomCode)
	if err != nil {
		return false
	}
	_, exists := s.rooms[intCode]
	return exists
}

func (s *Service) StartRoomByCode(roomCode int) (bool, error) {
	s.mu.RLock()
	room, exists := s.rooms[roomCode]
	s.mu.RUnlock()
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

	msg := util.Message{Header: "START", Body: ""}
	for _, c := range clients {
		_ = c.WriteJSON(msg)
	}

	return true, nil
}


// Room code is verified before this function is called.
func (s *Service) JoinRoomLocked(roomCode int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	room, _ := s.rooms[roomCode]
	if room.Started {
		return false
	}
	return true
}

func (s *Service) JoinRoom(codeStr string) (bool, error) {
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return false, fmt.Errorf("room code not numeric")
	}

	if !s.VerifyCode(codeStr) {
		return false, fmt.Errorf("invalid room code")
	}

	joined := s.JoinRoomLocked(code)
	return joined, nil
}

func (s *Service) StartRoom(codeStr string) (bool, error) {
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		return false, fmt.Errorf("room code not numeric")
	}
	return s.StartRoomByCode(code)
}


func (s *Service) RegisterConn(roomCode int, conn *websocket.Conn) (alreadyStarted bool, err error) {
	s.mu.RLock()
	room, exists := s.rooms[roomCode]
	s.mu.RUnlock()
	if !exists {
		return false, fmt.Errorf("room not found")
	}

	room.mu.Lock()
	room.clients[conn] = struct{}{}
	alreadyStarted = room.Started
	room.mu.Unlock()

	return alreadyStarted, nil
}

func (s *Service) UnregisterConn(roomCode int, conn *websocket.Conn) {
	s.mu.RLock()
	room, exists := s.rooms[roomCode]
	s.mu.RUnlock()
	if !exists {
		return
	}

	room.mu.Lock()
	delete(room.clients, conn)
	room.mu.Unlock()
}

