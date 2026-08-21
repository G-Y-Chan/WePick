package main

import (
	"backend/internal/config"
	"backend/internal/infra"
	"backend/internal/places"
	"backend/internal/room"
	api "backend/internal/transport/http"
	"backend/internal/transport/http/middleware"
	"backend/internal/transport/ws"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	_ = godotenv.Load()

	// Load configuration
	cfg := config.LoadEnv()

	// Initialize slog with level based on environment
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if cfg.Env == "development" {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))

	slog.Info("starting server", "env", cfg.Env, "port", cfg.Port)

	// Initialize dependencies
	rdb := infra.NewRedisClient(cfg.Redis)
	placesClient := places.NewPlacesClient(cfg.GooglePlaces.APIKey)
	hub := ws.NewHub()

	roomRepo := room.NewRoomRepository(rdb)
	roomManager := room.NewRoomManager(roomRepo, placesClient, hub)

	// Start the event listener in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go roomManager.StartEventListener(ctx)

	// Set up CORS allowlist
	middleware.AllowedOrigins = cfg.AllowedOrigins

	// Build the HTTP server
	srv := &api.Server{RoomManager: roomManager}
	mux := http.NewServeMux()

	// Health check endpoints (always available)
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintln(w, "unhealthy")
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", healthHandler)

	// Room endpoints
	mux.HandleFunc("GET /rooms", srv.GetRoomCode)
	mux.HandleFunc("POST /rooms/{code}/join", srv.HandleRoomJoin)
	mux.HandleFunc("POST /rooms/{code}/start", srv.HandleRoomStart)
	mux.HandleFunc("GET /rooms/{code}/cards", srv.HandleGetCardData)
	mux.HandleFunc("GET /image", srv.HandleGetImage)

	// WebSocket endpoint
	wsHandler := ws.NewHandler(roomManager, hub)
	mux.HandleFunc("/ws", wsHandler.HandleRoomWS)

	// Dev endpoints - conditionally registered based on environment
	if cfg.Env == "development" {
		slog.Info("registering development-only endpoints")
		mux.HandleFunc("/test", srv.Test)
		mux.HandleFunc("/headers", srv.Headers)
		mux.HandleFunc("/post-email", srv.PostEmail)
	}

	// Wrap with CORS middleware
	handler := middleware.WithCORSHandler(mux)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16, // 64 KB
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer shutdownCancel()

		cancel() // stop the event listener

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server shutdown error", "error", err)
		}
	}()

	slog.Info("listening", "addr", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}