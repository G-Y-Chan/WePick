package main

import (
    "net/http"
	"backend/room"
	"backend/middleware"
	"backend/handlers"
)

func main() {
	roomService := room.NewService(1_000_000)
	s := &handlers.Server{
		RoomService: roomService,
	}

	// Test endpoints
	http.HandleFunc("/test", middleware.WithCORS(s.Test))
	http.HandleFunc("/headers", middleware.WithCORS(s.Headers))
	http.HandleFunc("/post-email", middleware.WithCORS(s.PostEmail))

	// Actual endpoints
	http.HandleFunc("/get-room-code", middleware.WithCORS(s.GetRoomCode))
	http.HandleFunc("/join-room", middleware.WithCORS(s.HandleRoomJoin))
	http.HandleFunc("/start-room", middleware.WithCORS(s.HandleRoomStart))

	// WebSocket endpoint
	http.HandleFunc("/ws", s.HandleRoomWS)

	http.ListenAndServe(":8090", nil)
}