// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/skys-mission/key-agent/internal/common"
	"github.com/skys-mission/key-agent/internal/storage"
)

type ctxKey string

const (
	// TokenKey is the context key for token metadata.
	TokenKey ctxKey = "token"
)

// Auth provides Bearer Token authentication middleware.
type Auth struct {
	store storage.Store
}

// NewAuth creates a new auth middleware.
func NewAuth(store storage.Store) *Auth {
	return &Auth{store: store}
}

// RequireAuth is a middleware that validates Bearer tokens.
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid token")
			return
		}

		tokenHash := common.HashToken(token)
		meta, err := a.store.GetToken(r.Context(), tokenHash)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
			return
		}

		if meta.ExpiresAt != nil && meta.ExpiresAt.Before(time.Now()) {
			writeAuthError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token has expired")
			return
		}

		ctx := context.WithValue(r.Context(), TokenKey, meta)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractToken extracts the Bearer token from the Authorization header.
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return parts[1]
}

// writeAuthError writes a JSON error response.
func writeAuthError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
