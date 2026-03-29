// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/skys-mission/key-agent/internal/storage"
)

// SecretHandler handles secret operations.
type SecretHandler struct {
	store storage.Store
}

// NewSecretHandler creates a new secret handler.
func NewSecretHandler(store storage.Store) *SecretHandler {
	return &SecretHandler{store: store}
}

// ServeHTTP routes secret requests.
func (h *SecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			h.list(w, r)
		default:
			_ = writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
		return
	}

	key := strings.TrimPrefix(path, "/")
	if key == "" {
		_ = writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Key is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, key)
	case http.MethodPut:
		h.set(w, r, key)
	case http.MethodDelete:
		h.delete(w, r, key)
	default:
		_ = writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// list handles GET /api/v1/secrets.
func (h *SecretHandler) list(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")

	keys, err := h.store.ListSecret(r.Context(), prefix)
	if err != nil {
		_ = writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list secrets")
		return
	}

	if keys == nil {
		keys = []string{}
	}

	_ = writeJSON(w, http.StatusOK, map[string]interface{}{"keys": keys})
}

// get handles GET /api/v1/secrets/{key}.
func (h *SecretHandler) get(w http.ResponseWriter, r *http.Request, key string) {
	entry, err := h.store.GetSecret(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			_ = writeError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
			return
		}
		_ = writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get secret")
		return
	}

	_ = writeJSON(w, http.StatusOK, entry)
}

// SetSecretRequest represents the request body for setting a secret.
type SetSecretRequest struct {
	Value    string             `json:"value"`
	Type     storage.SecretType `json:"type"`
	Metadata storage.Metadata   `json:"metadata,omitempty"`
}

// set handles PUT /api/v1/secrets/{key}.
func (h *SecretHandler) set(w http.ResponseWriter, r *http.Request, key string) {
	var req SetSecretRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if req.Value == "" {
		_ = writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Value is required")
		return
	}

	if req.Type == "" {
		_ = writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Type is required")
		return
	}

	// Validate secret type
	validTypes := map[storage.SecretType]bool{
		storage.SecretTypePassword:    true,
		storage.SecretTypeAPIKey:      true,
		storage.SecretTypeCertificate: true,
		storage.SecretTypePrivateKey:  true,
		storage.SecretTypeToken:       true,
		storage.SecretTypeOther:       true,
	}
	if !validTypes[req.Type] {
		_ = writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid secret type")
		return
	}

	// Check if entry exists
	existing, err := h.store.GetSecret(r.Context(), key)
	isNew := err == storage.ErrNotFound

	entry := &storage.SecretEntry{
		Entry: storage.Entry{
			Key:      key,
			Value:    req.Value,
			Metadata: req.Metadata,
		},
		Type: req.Type,
	}
	if existing != nil {
		entry.CreatedAt = existing.CreatedAt
		entry.Version = existing.Version
	}

	if err := h.store.SetSecret(r.Context(), entry); err != nil {
		_ = writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set secret")
		return
	}

	// Fetch the updated entry
	entry, _ = h.store.GetSecret(r.Context(), key)

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}

	_ = writeJSON(w, status, entry)
}

// delete handles DELETE /api/v1/secrets/{key}.
func (h *SecretHandler) delete(w http.ResponseWriter, r *http.Request, key string) {
	err := h.store.DeleteSecret(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			_ = writeError(w, http.StatusNotFound, "NOT_FOUND", "Secret not found")
			return
		}
		_ = writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete secret")
		return
	}

	writeNoContent(w)
}
