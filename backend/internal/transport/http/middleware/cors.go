package middleware

import (
	"net/http"
	"strings"
)

// AllowedOrigins is injected at startup by main.go.
// It's a reference so we can swap it without rebuilding middleware.
var AllowedOrigins []string

// isAllowed returns true if the origin is allowed.
func isAllowed(origin string) bool {
	for _, allowed := range AllowedOrigins {
		if allowed == origin {
			return true
		}
		// Allow wildcard suffixes (e.g. ".exp.direct", ".expo.dev")
		if strings.HasPrefix(allowed, ".") && strings.HasSuffix(origin, allowed) {
			return true
		}
		// Allow wildcard prefixes (e.g. "*.ngrok-free.app")
		if strings.HasPrefix(allowed, ".*") && strings.HasSuffix(origin, allowed[1:]) {
			return true
		}
	}
	return false
}

// WithCORS wraps an http.HandlerFunc with CORS headers based on the allowlist.
func WithCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowed(origin) && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// WithCORSHandler wraps an http.Handler with CORS headers.
func WithCORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowed(origin) && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}