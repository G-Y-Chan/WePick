package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"backend/config"
	"backend/internal/adapters/places"
	"backend/internal/adapters/redisrepo"
	"backend/internal/adapters/wshub"
	"backend/internal/domain/room"
	"backend/internal/platform/redisconn"
	httptransport "backend/internal/transport/http"
	"backend/internal/transport/http/middleware"
	wstransport "backend/internal/transport/ws"
)

// allowedOrigins is the single authoritative CORS / WebSocket origin
// allow-list. It restores the policy that already existed — unused — in the
// legacy backend/middleware/cors.go and replaces the wide-open rs/cors
// configuration that shipped in the old main.go.
var allowedOrigins = []string{
	"http://localhost:8081",
	"http://localhost:19006",
	"*.exp.direct",
	"*.expo.dev",
	"*.ngrok-free.app",
}

func main() {
	_ = godotenv.Load()

	cfg := config.LoadEnv()
	if cfg.Port == "" {
		cfg.Port = "8090"
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Composition root: build adapters, wire the domain service, then hand the
	// service to the transport layer. This is the only file in the module that
	// should import every one of these packages directly.
	rdb := redisconn.New(cfg.Redis)
	repo := redisrepo.New(rdb)
	placesClient := places.New(cfg.GooglePlaces.APIKey, nil)
	hub := wshub.NewHub(logger)

	svc := room.NewService(repo, placesClient, hub)

	// Start the cross-process Redis pub/sub event listener. Phase 7 replaces
	// this Background context with the signal-derived root context and waits
	// for the goroutine during graceful shutdown.
	if runner, ok := svc.(room.EventLoopRunner); ok {
		go runner.StartEventListener(context.Background())
	}

	httpHandler := httptransport.NewHandler(svc, logger)
	wsHandler := wstransport.NewHandler(svc, allowedOrigins, logger)

	root := http.NewServeMux()
	root.Handle("/", httpHandler.Routes())
	root.HandleFunc("/ws", wsHandler.ServeWS)

	handler := middleware.RequestID()(
		middleware.Recover(logger)(
			middleware.AccessLog(logger)(
				middleware.CORS(allowedOrigins)(root),
			),
		),
	)

	addr := ":" + cfg.Port
	logger.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}
