//go:build dev

package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// Test handler — only compiled with `go build -tags dev`.
func (s *Server) Test(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	testMessage := "This is a placeholder body."
	m := Message{
		Header: "testing",
		Body:   &testMessage,
	}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Headers handler — only compiled with `go build -tags dev`.
func (s *Server) Headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			_, _ = fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

// PostEmail handler — only compiled with `go build -tags dev`.
func (s *Server) PostEmail(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	var email string
	err := json.NewDecoder(req.Body).Decode(&email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slog.Info("email received (dev only)", "email", email)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}