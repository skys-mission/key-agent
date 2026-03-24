// Package server provides HTTP routing.
package server

import (
	"net/http"

	"github.com/skys-mission/key-agent/internal/mcp"
	"github.com/skys-mission/key-agent/internal/server/handlers"
	"github.com/skys-mission/key-agent/internal/server/middleware"
	"github.com/skys-mission/key-agent/internal/storage"
)

// Router creates the HTTP router with all handlers.
func Router(store storage.Store, mcpServer *mcp.Server, version string) http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth required)
	healthHandler := handlers.NewHealthHandler(version)
	mux.Handle("/health", healthHandler)

	// API handlers
	kvHandler := handlers.NewKVHandler(store)
	secretHandler := handlers.NewSecretHandler(store)
	tokenHandler := handlers.NewTokenHandler(store)

	// Auth middleware
	auth := middleware.NewAuth(store)

	// Protected routes
	mux.Handle("/api/v1/kv", auth.RequireAuth(http.StripPrefix("/api/v1/kv", kvHandler)))
	mux.Handle("/api/v1/kv/", auth.RequireAuth(http.StripPrefix("/api/v1/kv", kvHandler)))
	mux.Handle("/api/v1/secrets", auth.RequireAuth(http.StripPrefix("/api/v1/secrets", secretHandler)))
	mux.Handle("/api/v1/secrets/", auth.RequireAuth(http.StripPrefix("/api/v1/secrets", secretHandler)))
	mux.Handle("/api/v1/token", auth.RequireAuth(tokenHandler))
	mux.Handle("/api/v1/token/", auth.RequireAuth(tokenHandler))

	// MCP endpoint (protected by auth)
	if mcpServer != nil {
		mcpHandler := mcpServer.Handler()
		mux.Handle("/mcp", auth.RequireAuth(mcpHandler))
		mux.Handle("/mcp/", auth.RequireAuth(mcpHandler))
	}

	// Logging middleware
	return middleware.Logging(mux)
}
