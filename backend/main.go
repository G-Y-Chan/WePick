package main

import (
	"backend/config"
	"backend/handlers"
	"backend/infra"
	"backend/room"
	"backend/services"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Initialize Redis client and repositories
	cfg := config.LoadEnv()
	rdb := infra.NewRedisClient(cfg.Redis)
	roomRepo := room.NewRoomRepository(rdb)
	placesClient := services.NewPlacesClient(cfg.GooglePlaces.APIKey)
	roomManager := room.NewRoomManager(1_000_000, roomRepo, placesClient)

	go roomManager.StartEventListener()

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
	mux.HandleFunc("GET /rooms", s.GetRoomCode)
	mux.HandleFunc("POST /rooms/{code}/join", s.HandleRoomJoin)
	mux.HandleFunc("POST /rooms/{code}/start", s.HandleRoomStart)
	mux.HandleFunc("GET /rooms/{code}/cards", s.HandleGetCardData)

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

	addr := ":" + cfg.Port

	_ = http.ListenAndServe(addr, handler)
}
