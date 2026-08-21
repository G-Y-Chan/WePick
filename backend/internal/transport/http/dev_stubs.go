//go:build !dev

package api

import (
	"net/http"
)

// Test is a no-op in production builds.
func (s *Server) Test(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "not available", http.StatusNotFound)
}

// Headers is a no-op in production builds.
func (s *Server) Headers(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "not available", http.StatusNotFound)
}

// PostEmail is a no-op in production builds.
func (s *Server) PostEmail(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "not available", http.StatusNotFound)
}