package room

import (
	"sync"
	"github.com/gorilla/websocket"
)

type Room struct {
    Code      string
    Started   bool

	clients   map[*websocket.Conn]struct{}
	mu        sync.RWMutex
}
