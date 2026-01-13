package main

import (
    "net/http"
	"backend/room"
	"backend/middleware"
)

type Server struct {
	roomService *room.Service
}

func main() {
	roomService := room.NewService(1_000_000)
	s := &Server{
		roomService: roomService,
	}

	// Test endpoints
	http.HandleFunc("/test", middleware.WithCORS(s.test))
	http.HandleFunc("/headers", middleware.WithCORS(s.headers))
	http.HandleFunc("/post-email", middleware.WithCORS(s.postEmail))

	// Actual endpoints
	http.HandleFunc("/get-room-code", middleware.WithCORS(s.getRoomCode))
	http.HandleFunc("/verify-room-code", middleware.WithCORS(s.verifyRoomCode))
	http.HandleFunc("/start-room", middleware.WithCORS(s.verifyStart))
    http.ListenAndServe(":8090", nil)
}