package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a middleware that applies the single authoritative CORS policy.
// It replaces both the legacy middleware.WithCORS (dead code) and the
// wide-open rs/cors configuration that previously allowed every origin.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if origin != "" && originAllowed(origin, allowedOrigins) {
				// IMPORTANT: Access-Control-Allow-Origin must match the request
				// origin exactly (not a wildcard) when the client uses cookies.
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}

			// Preflight requests must always short-circuit: no route handler
			// should ever be invoked for OPTIONS.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// originAllowed reports whether origin matches one of the allow-list entries.
// Exact entries are compared verbatim; entries beginning with "*." are treated
// as suffix patterns so subdomains like "https://abc.expo.dev" match "*.expo.dev".
func originAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(allowed, "*.") {
			suffix := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, suffix) {
				return true
			}
			continue
		}

		if origin == allowed {
			return true
		}
	}
	return false
}
