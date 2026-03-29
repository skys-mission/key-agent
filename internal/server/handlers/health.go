// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	version string
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

// ServeHTTP handles GET /health.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"version": h.version,
	})
}
