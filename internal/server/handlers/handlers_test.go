package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/skys-mission/key-agent/internal/storage"
)

// mockStore implements storage.Store for testing
type mockStore struct {
	kvEntries     map[string]*storage.Entry
	secretEntries map[string]*storage.SecretEntry
	tokens        map[string]*storage.TokenMeta
}

func newMockStore() *mockStore {
	return &mockStore{
		kvEntries:     make(map[string]*storage.Entry),
		secretEntries: make(map[string]*storage.SecretEntry),
		tokens:        make(map[string]*storage.TokenMeta),
	}
}

func (m *mockStore) GetKV(ctx context.Context, key string) (*storage.Entry, error) {
	if e, ok := m.kvEntries[key]; ok {
		return e, nil
	}
	return nil, storage.ErrNotFound
}

func (m *mockStore) SetKV(ctx context.Context, entry *storage.Entry) error {
	now := time.Now()
	if e, ok := m.kvEntries[entry.Key]; ok {
		entry.CreatedAt = e.CreatedAt
	} else {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	m.kvEntries[entry.Key] = entry
	return nil
}

func (m *mockStore) DeleteKV(ctx context.Context, key string) error {
	if _, ok := m.kvEntries[key]; !ok {
		return storage.ErrNotFound
	}
	delete(m.kvEntries, key)
	return nil
}

func (m *mockStore) ListKV(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for k := range m.kvEntries {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockStore) GetSecret(ctx context.Context, key string) (*storage.SecretEntry, error) {
	if e, ok := m.secretEntries[key]; ok {
		return e, nil
	}
	return nil, storage.ErrNotFound
}

func (m *mockStore) SetSecret(ctx context.Context, entry *storage.SecretEntry) error {
	m.secretEntries[entry.Key] = entry
	return nil
}

func (m *mockStore) DeleteSecret(ctx context.Context, key string) error {
	delete(m.secretEntries, key)
	return nil
}

func (m *mockStore) ListSecret(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for k := range m.secretEntries {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockStore) GetToken(ctx context.Context, tokenHash string) (*storage.TokenMeta, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, storage.ErrNotFound
}

func (m *mockStore) SetToken(ctx context.Context, tokenHash string, meta *storage.TokenMeta) error {
	m.tokens[tokenHash] = meta
	return nil
}

func (m *mockStore) DeleteToken(ctx context.Context, tokenHash string) error {
	delete(m.tokens, tokenHash)
	return nil
}

func (m *mockStore) ListTokens(ctx context.Context) ([]string, error) {
	var hashes []string
	for h := range m.tokens {
		hashes = append(hashes, h)
	}
	return hashes, nil
}

func (m *mockStore) Close() error { return nil }

func TestHealthHandler(t *testing.T) {
	handler := NewHealthHandler("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("HealthHandler status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if resp["status"] != "healthy" {
		t.Errorf("HealthHandler status = %v, want healthy", resp["status"])
	}
	if resp["version"] != "1.0.0" {
		t.Errorf("HealthHandler version = %v, want 1.0.0", resp["version"])
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler("1.0.0")

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("HealthHandler status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestKVHandler_List(t *testing.T) {
	store := newMockStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "key1", Value: "v1"})
	store.SetKV(context.Background(), &storage.Entry{Key: "key2", Value: "v2"})

	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kv", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("KVHandler list status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string][]string
	json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp["keys"]) != 2 {
		t.Errorf("KVHandler list keys count = %d, want 2", len(resp["keys"]))
	}
}

func TestKVHandler_Get(t *testing.T) {
	store := newMockStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "test-key", Value: "test-value"})

	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kv/test-key", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("KVHandler get status = %d, want %d", rec.Code, http.StatusOK)
	}

	var entry storage.Entry
	json.NewDecoder(rec.Body).Decode(&entry)

	if entry.Key != "test-key" {
		t.Errorf("KVHandler get key = %v, want test-key", entry.Key)
	}
	if entry.Value != "test-value" {
		t.Errorf("KVHandler get value = %v, want test-value", entry.Value)
	}
}

func TestKVHandler_GetNotFound(t *testing.T) {
	store := newMockStore()
	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/kv/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("KVHandler get status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestKVHandler_Set(t *testing.T) {
	store := newMockStore()
	handler := NewKVHandler(store)

	body := bytes.NewBufferString(`{"value":"new-value"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/kv/new-key", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("KVHandler set status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var entry storage.Entry
	json.NewDecoder(rec.Body).Decode(&entry)

	if entry.Key != "new-key" {
		t.Errorf("KVHandler set key = %v, want new-key", entry.Key)
	}
	if entry.Value != "new-value" {
		t.Errorf("KVHandler set value = %v, want new-value", entry.Value)
	}
}

func TestKVHandler_SetExisting(t *testing.T) {
	store := newMockStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "existing-key", Value: "old-value"})

	handler := NewKVHandler(store)

	body := bytes.NewBufferString(`{"value":"updated-value"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/kv/existing-key", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("KVHandler set status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestKVHandler_SetEmptyValue(t *testing.T) {
	store := newMockStore()
	handler := NewKVHandler(store)

	body := bytes.NewBufferString(`{"value":""}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/kv/key", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("KVHandler set status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestKVHandler_Delete(t *testing.T) {
	store := newMockStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "delete-key", Value: "value"})

	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/kv/delete-key", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("KVHandler delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestKVHandler_DeleteNotFound(t *testing.T) {
	store := newMockStore()
	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/kv/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("KVHandler delete status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestKVHandler_MethodNotAllowed(t *testing.T) {
	store := newMockStore()
	handler := NewKVHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/kv/key", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("KVHandler status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}
