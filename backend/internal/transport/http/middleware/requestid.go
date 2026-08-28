package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

type requestIDKey struct{}

// RequestID returns a middleware that reuses an inbound X-Request-ID header
// when present and otherwise generates a new request ID, then propagates it
// through the request context and echoes it on the response.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = newRequestID()
			}

			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequestIDFromContext returns the request ID injected by RequestID, or "" if
// the middleware was not applied.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err == nil {
		return hex.EncodeToString(b)
	}

	// crypto/rand should not fail on a supported platform, but fall back to a
	// timestamp so request logging still has a unique-ish value.
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
