package api

import (
	"backend/internal/room"
	"backend/internal/transport/ws"
)

// Server holds dependencies for HTTP handlers.
type Server struct {
	RoomManager *room.RoomManager
}

func NewServer(rm *room.RoomManager) *Server {
	return &Server{RoomManager: rm}
}

// WsHandler returns the WebSocket handler wired with the Hub.
func WsHandler(rm *room.RoomManager, hub *ws.Hub) *ws.Handler {
	return ws.NewHandler(rm, hub)
}