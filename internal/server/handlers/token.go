// Package handlers provides HTTP request handlers.
package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/skys-mission/key-agent/internal/common"
	"github.com/skys-mission/key-agent/internal/storage"
)

// TokenHandler handles token operations.
type TokenHandler struct {
	store storage.Store
}

// NewTokenHandler creates a new token handler.
func NewTokenHandler(store storage.Store) *TokenHandler {
	return &TokenHandler{store: store}
}

// ServeHTTP routes token requests.
func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/token")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodPost:
			h.create(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
		return
	}

	writeError(w, http.StatusNotFound, "NOT_FOUND", "Not found")
}

// CreateTokenRequest represents the request body for creating a token.
type CreateTokenRequest struct {
	Name      string `json:"name"`
	Type      string `json:"type,omitempty"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

// CreateTokenResponse represents the response for creating a token.
type CreateTokenResponse struct {
	Token     string     `json:"token"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// create handles POST /api/v1/token.
func (h *TokenHandler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Name is required")
		return
	}

	// Validate token type
	tokenType := req.Type
	if tokenType == "" {
		tokenType = "client"
	}
	if tokenType != "client" && tokenType != "mcp" {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid token type")
		return
	}

	// Generate token
	token, err := generateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token")
		return
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresIn != "" {
		d, err := parseDuration(req.ExpiresIn)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid duration format")
			return
		}
		t := time.Now().Add(d)
		expiresAt = &t
	}

	// Store token metadata
	now := time.Now()
	meta := &storage.TokenMeta{
		Name:      req.Name,
		Type:      tokenType,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	tokenHash := common.HashToken(token)
	if err := h.store.SetToken(r.Context(), tokenHash, meta); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to store token")
		return
	}

	writeJSON(w, http.StatusCreated, CreateTokenResponse{
		Token:     token,
		Name:      req.Name,
		Type:      tokenType,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	})
}

// generateToken generates a secure random token.
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ka_" + hex.EncodeToString(bytes), nil
}

var errInvalidDuration = errors.New("invalid duration format")

// parseDuration parses a duration string like "24h", "30d".
func parseDuration(s string) (time.Duration, error) {
	// Handle days
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days := 0
		for _, c := range daysStr {
			if c < '0' || c > '9' {
				return 0, errInvalidDuration
			}
			days = days*10 + int(c-'0')
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
