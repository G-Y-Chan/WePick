package middleware

import (
	"net/http"
	"strings"
)

func WithCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow:
		// - localhost dev
		// - Expo web dev hosts (*.exp.direct)
		allowed := origin == "http://localhost:8081" ||
			origin == "http://localhost:19006" ||
			strings.HasSuffix(origin, ".exp.direct") ||
			strings.HasSuffix(origin, ".expo.dev") ||
            strings.HasSuffix(origin, ".ngrok-free.app")


		if allowed && origin != "" {
			// IMPORTANT: must match the request Origin exactly (not a different one)
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			// Only set this if you actually use cookies/credentials:
			// w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
