// Package handlers provides HTTP request handlers.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/skys-mission/key-agent/internal/storage"
)

// KVHandler handles KV operations.
type KVHandler struct {
	store storage.Store
}

// NewKVHandler creates a new KV handler.
func NewKVHandler(store storage.Store) *KVHandler {
	return &KVHandler{store: store}
}

// ServeHTTP routes KV requests.
func (h *KVHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/kv")

	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			h.list(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		}
		return
	}

	key := strings.TrimPrefix(path, "/")
	if key == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Key is required")
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
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// list handles GET /api/v1/kv.
func (h *KVHandler) list(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")

	keys, err := h.store.ListKV(r.Context(), prefix)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list keys")
		return
	}

	if keys == nil {
		keys = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"keys": keys})
}

// get handles GET /api/v1/kv/{key}.
func (h *KVHandler) get(w http.ResponseWriter, r *http.Request, key string) {
	entry, err := h.store.GetKV(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get key")
		return
	}

	writeJSON(w, http.StatusOK, entry)
}

// SetKVRequest represents the request body for setting a KV entry.
type SetKVRequest struct {
	Value    string           `json:"value"`
	Metadata storage.Metadata `json:"metadata,omitempty"`
}

// set handles PUT /api/v1/kv/{key}.
func (h *KVHandler) set(w http.ResponseWriter, r *http.Request, key string) {
	var req SetKVRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if req.Value == "" {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Value is required")
		return
	}

	// Check if entry exists
	existing, err := h.store.GetKV(r.Context(), key)
	isNew := err == storage.ErrNotFound

	entry := &storage.Entry{
		Key:      key,
		Value:    req.Value,
		Metadata: req.Metadata,
	}
	if existing != nil {
		entry.CreatedAt = existing.CreatedAt
		entry.Version = existing.Version
	}

	if err := h.store.SetKV(r.Context(), entry); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to set key")
		return
	}

	// Fetch the updated entry
	entry, _ = h.store.GetKV(r.Context(), key)

	status := http.StatusOK
	if isNew {
		status = http.StatusCreated
	}

	writeJSON(w, status, entry)
}

// delete handles DELETE /api/v1/kv/{key}.
func (h *KVHandler) delete(w http.ResponseWriter, r *http.Request, key string) {
	err := h.store.DeleteKV(r.Context(), key)
	if err != nil {
		if err == storage.ErrNotFound {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Key not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete key")
		return
	}

	writeNoContent(w)
}
