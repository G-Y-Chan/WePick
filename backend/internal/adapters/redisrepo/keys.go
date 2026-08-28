package redisrepo

import (
	"fmt"

	"backend/internal/domain/room"
)

// Key-naming functions carried over from room/room_repository.go unchanged.
// Redis key format and hash field names are preserved to keep live data readable.

func roomKey(code room.Code) string {
	return "room:" + code.String()
}

func voteKey(roomCode room.Code) string {
	return fmt.Sprintf("votes:%s", roomCode.String())
}
