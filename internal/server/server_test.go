package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/skys-mission/key-agent/internal/storage"
)

func TestNew(t *testing.T) {
	handler := http.NewServeMux()
	s := New("127.0.0.1:8080", handler)

	if s.addr != "127.0.0.1:8080" {
		t.Errorf("Server addr = %v, want 127.0.0.1:8080", s.addr)
	}
	if s.handler == nil {
		t.Error("Server handler should not be nil")
	}
}

func TestShutdown_NilServer(t *testing.T) {
	s := &Server{}
	err := s.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() with nil server error = %v", err)
	}
}

func TestRouter_HealthEndpoint(t *testing.T) {
	store := newMockRouterStore()
	router := Router(store, nil, "1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Health endpoint status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRouter_AuthRequired(t *testing.T) {
	store := newMockRouterStore()
	router := Router(store, nil, "1.0.0")

	// KV endpoint should require auth
	req := httptest.NewRequest(http.MethodGet, "/api/v1/kv", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("KV endpoint without auth status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// mockRouterStore implements storage.Store for routing tests
type mockRouterStore struct{}

func newMockRouterStore() *mockRouterStore {
	return &mockRouterStore{}
}

func (m *mockRouterStore) GetKV(ctx context.Context, key string) (*storage.Entry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRouterStore) SetKV(ctx context.Context, entry *storage.Entry) error {
	return nil
}
func (m *mockRouterStore) DeleteKV(ctx context.Context, key string) error {
	return nil
}
func (m *mockRouterStore) ListKV(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}
func (m *mockRouterStore) GetSecret(ctx context.Context, key string) (*storage.SecretEntry, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRouterStore) SetSecret(ctx context.Context, entry *storage.SecretEntry) error {
	return nil
}
func (m *mockRouterStore) DeleteSecret(ctx context.Context, key string) error {
	return nil
}
func (m *mockRouterStore) ListSecret(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}
func (m *mockRouterStore) GetToken(ctx context.Context, tokenHash string) (*storage.TokenMeta, error) {
	return nil, storage.ErrNotFound
}
func (m *mockRouterStore) SetToken(ctx context.Context, tokenHash string, meta *storage.TokenMeta) error {
	return nil
}
func (m *mockRouterStore) DeleteToken(ctx context.Context, tokenHash string) error {
	return nil
}
func (m *mockRouterStore) ListTokens(ctx context.Context) ([]string, error) {
	return nil, nil
}
func (m *mockRouterStore) Close() error { return nil }
