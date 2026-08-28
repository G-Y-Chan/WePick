package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog returns a middleware that emits one structured log line per HTTP
// request. It records the actual response status even when a downstream
// middleware or handler short-circuits.
func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration", time.Since(start),
				"remote_addr", r.RemoteAddr,
				"request_id", RequestIDFromContext(r.Context()),
			)
		})
	}
}
