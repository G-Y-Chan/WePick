package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"backend/config"
	"backend/internal/adapters/places"
	"backend/internal/adapters/redisrepo"
	"backend/internal/adapters/wshub"
	"backend/internal/domain/room"
	"backend/internal/platform/logging"
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

	logger := logging.New(slog.LevelInfo)

	// ---- Composition root ------------------------------------------------
	rdb := redisconn.New(cfg.Redis)
	repo := redisrepo.New(rdb)
	placesClient := places.New(cfg.GooglePlaces.APIKey, nil)
	hub := wshub.NewHub(logger)

	svc := room.NewService(repo, placesClient, hub, room.WithLogger(logger))

	httpHandler := httptransport.NewHandler(svc, logger).
		WithHealthChecker(func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		})
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

	// ---- Signal-aware lifecycle ------------------------------------------
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	// Start the cross-process Redis pub/sub event listener with a cancellable
	// context derived from the signal-aware root so the goroutine can be
	// stopped cleanly on shutdown.
	if runner, ok := svc.(room.EventLoopRunner); ok {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.StartEventListener(ctx)
		}()
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Run the HTTP server in a goroutine so we can listen for shutdown signals
	// on the main goroutine.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	// Block until a signal arrives or the server dies unexpectedly.
	select {
	case <-ctx.Done():
		logger.Info("shutting down gracefully", "reason", ctx.Err())
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "err", err)
		}
	}

	// Give in-flight requests a bounded window to complete.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}

	// Trigger the event-listener goroutine stop and wait for it.
	stop()
	wg.Wait()

	logger.Info("server stopped")
}
