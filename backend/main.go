package main

import (
    "net/http"
	"backend/room"
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
	http.HandleFunc("/test", s.test)
	http.HandleFunc("/headers", s.headers)
	http.HandleFunc("/post-email", s.postEmail)

	// Actual endpoints
	http.HandleFunc("/get-room-code", s.getRoomCode)
	http.HandleFunc("/verify-room-code", s.verifyRoomCode)
	http.HandleFunc("/start-room", s.verifyStart)
    http.ListenAndServe(":8090", nil)
}