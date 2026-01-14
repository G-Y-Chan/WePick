package room

import (
	"math/rand"
	"strconv"
	"sync"
)

type Room struct {
    Code      int
    Started   bool
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