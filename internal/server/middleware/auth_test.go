package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skys-mission/key-agent/internal/common"
	"github.com/skys-mission/key-agent/internal/storage"
)

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantToken  string
	}{
		{"valid bearer token", "Bearer mytoken123", "mytoken123"},
		{"lowercase bearer", "bearer mytoken123", "mytoken123"},
		{"no auth header", "", ""},
		{"wrong scheme", "Basic mytoken123", ""},
		{"malformed header", "Bearer", ""},
		{"bearer with leading space in token", "Bearer  mytoken", " mytoken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			got := extractToken(req)
			if got != tt.wantToken {
				t.Errorf("extractToken() = %v, want %v", got, tt.wantToken)
			}
		})
	}
}

// mockAuthStore implements storage.Store for auth testing
type mockAuthStore struct {
	tokens map[string]*storage.TokenMeta
}

func newMockAuthStore() *mockAuthStore {
	return &mockAuthStore{tokens: make(map[string]*storage.TokenMeta)}
}

func (m *mockAuthStore) GetKV(ctx context.Context, key string) (*storage.Entry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockAuthStore) SetKV(ctx context.Context, entry *storage.Entry) error { return nil }
func (m *mockAuthStore) DeleteKV(ctx context.Context, key string) error        { return nil }
func (m *mockAuthStore) ListKV(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}
func (m *mockAuthStore) GetSecret(ctx context.Context, key string) (*storage.SecretEntry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockAuthStore) SetSecret(ctx context.Context, entry *storage.SecretEntry) error {
	return nil
}
func (m *mockAuthStore) DeleteSecret(ctx context.Context, key string) error { return nil }
func (m *mockAuthStore) ListSecret(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}
func (m *mockAuthStore) GetToken(ctx context.Context, tokenHash string) (*storage.TokenMeta, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, storage.ErrNotFound
}
func (m *mockAuthStore) SetToken(ctx context.Context, tokenHash string, meta *storage.TokenMeta) error {
	m.tokens[tokenHash] = meta
	return nil
}
func (m *mockAuthStore) DeleteToken(ctx context.Context, tokenHash string) error {
	delete(m.tokens, tokenHash)
	return nil
}
func (m *mockAuthStore) ListTokens(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockAuthStore) Close() error { return nil }

func TestRequireAuth(t *testing.T) {
	// Setup store with a valid token
	store := newMockAuthStore()
	token := "valid-token-123"
	tokenHash := common.HashToken(token)
	store.SetToken(context.Background(), tokenHash, &storage.TokenMeta{
		Name:      "test",
		Type:      "client",
		CreatedAt: time.Now(),
	})

	// Create auth middleware
	auth := NewAuth(store)

	// Create a test handler that just returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"valid token", "Bearer " + token, http.StatusOK},
		{"invalid token", "Bearer invalid-token", http.StatusUnauthorized},
		{"no auth header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic " + token, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			auth.RequireAuth(testHandler).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("RequireAuth() status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	store := newMockAuthStore()
	token := "expired-token"
	tokenHash := common.HashToken(token)

	// Set an expired token
	pastTime := time.Now().Add(-24 * time.Hour)
	store.SetToken(context.Background(), tokenHash, &storage.TokenMeta{
		Name:      "expired",
		Type:      "client",
		CreatedAt: pastTime.Add(-24 * time.Hour),
		ExpiresAt: &pastTime, // Already expired
	})

	auth := NewAuth(store)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	auth.RequireAuth(testHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("RequireAuth() with expired token status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidTokenWithExpiry(t *testing.T) {
	store := newMockAuthStore()
	token := "valid-with-expiry"
	tokenHash := common.HashToken(token)

	// Set a valid token with future expiry
	futureTime := time.Now().Add(24 * time.Hour)
	store.SetToken(context.Background(), tokenHash, &storage.TokenMeta{
		Name:      "valid-expiry",
		Type:      "client",
		CreatedAt: time.Now(),
		ExpiresAt: &futureTime,
	})

	auth := NewAuth(store)
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	auth.RequireAuth(testHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("RequireAuth() with valid token status = %d, want %d", rec.Code, http.StatusOK)
	}
}
