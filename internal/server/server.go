// Package server provides the HTTP server implementation.
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/skys-mission/key-agent/internal/logger"
)

// Server represents the HTTP server.
type Server struct {
	addr    string
	handler http.Handler
	server  *http.Server
}

// New creates a new HTTP server.
func New(addr string, handler http.Handler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("HTTP server listening", "addr", s.addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}
