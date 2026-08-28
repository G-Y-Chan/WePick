package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Recover returns a middleware that converts an unrecovered handler panic into
// a clean HTTP 500 JSON response instead of an ungraceful connection reset.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w}

			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered in http handler",
						"panic", fmt.Sprint(recovered),
						"method", r.Method,
						"path", r.URL.Path,
					)

					// If the handler already committed a response we cannot
					// write a clean 500; the connection is already defined by
					// whatever was sent.
					if rec.wroteHeader {
						return
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(recoverErrorDTO{
						Header:  "Internal Server Error",
						Body:    "internal server error",
						Message: "",
					})
				}
			}()

			next.ServeHTTP(rec, r)
		})
	}
}

// statusRecorder tracks whether a response has been committed and which status
// code was written. It is shared by Recover and AccessLog.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}

// recoverErrorDTO mirrors the legacy ErrorResponse wire shape so even panic
// responses look like the rest of the API.
type recoverErrorDTO struct {
	Header  string `json:"Header"`
	Body    string `json:"Body"`
	Message string `json:"Message"`
}
