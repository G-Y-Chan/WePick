package main

import (
	"net/http"
	

	"backend/handlers"
	"backend/room"

	"github.com/rs/cors"
)

func main() {
	roomManager := room.NewRoomManager(1_000_000)
	s := &handlers.Server{
		RoomManager: roomManager,
	}

	// Register routes on a ServeMux (instead of using the global DefaultServeMux directly)
	mux := http.NewServeMux()

	// Test endpoints
	mux.HandleFunc("/test", s.Test)
	mux.HandleFunc("/headers", s.Headers)
	mux.HandleFunc("/post-email", s.PostEmail)

	// Actual endpoints
	mux.HandleFunc("/get-room-code", s.GetRoomCode)
	mux.HandleFunc("/join-room", s.HandleRoomJoin)
	mux.HandleFunc("/start-room", s.HandleRoomStart)
	mux.HandleFunc("/get-card-data", s.HandleGetCardData)

	// WebSocket endpoint (CORS isn’t enforced the same way for WS, but it’s fine to wrap anyway)
	mux.HandleFunc("/ws", s.HandleRoomWS)

	// Global CORS middleware
	c := cors.New(cors.Options{
		AllowOriginFunc: func(origin string) bool {
			return true
			/*
			return origin == "http://localhost:8081" ||
				origin == "http://localhost:19006" ||
				strings.HasSuffix(origin, ".exp.direct") ||
				strings.HasSuffix(origin, ".expo.dev") ||
				strings.HasSuffix(origin, ".ngrok-free.app")
			*/
		},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		// AllowedHeaders: []string{"Content-Type", "Authorization"},
		AllowedHeaders: []string{"*"},
		// AllowCredentials: true, // only if you use cookies / credentials
	})

	handler := c.Handler(mux)

	_ = http.ListenAndServe(":8090", handler)
}
