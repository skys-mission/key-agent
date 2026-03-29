package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/skys-mission/key-agent/internal/storage"
)

// mockMCPStore implements storage.Store for MCP testing
type mockMCPStore struct {
	kvEntries     map[string]*storage.Entry
	secretEntries map[string]*storage.SecretEntry
	tokens        map[string]*storage.TokenMeta
}

func newMockMCPStore() *mockMCPStore {
	return &mockMCPStore{
		kvEntries:     make(map[string]*storage.Entry),
		secretEntries: make(map[string]*storage.SecretEntry),
		tokens:        make(map[string]*storage.TokenMeta),
	}
}

func (m *mockMCPStore) GetKV(ctx context.Context, key string) (*storage.Entry, error) {
	if e, ok := m.kvEntries[key]; ok {
		return e, nil
	}
	return nil, storage.ErrNotFound
}
func (m *mockMCPStore) SetKV(ctx context.Context, entry *storage.Entry) error {
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
func (m *mockMCPStore) DeleteKV(ctx context.Context, key string) error {
	delete(m.kvEntries, key)
	return nil
}
func (m *mockMCPStore) ListKV(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for k := range m.kvEntries {
		keys = append(keys, k)
	}
	return keys, nil
}
func (m *mockMCPStore) GetSecret(ctx context.Context, key string) (*storage.SecretEntry, error) {
	if e, ok := m.secretEntries[key]; ok {
		return e, nil
	}
	return nil, storage.ErrNotFound
}
func (m *mockMCPStore) SetSecret(ctx context.Context, entry *storage.SecretEntry) error {
	m.secretEntries[entry.Key] = entry
	return nil
}
func (m *mockMCPStore) DeleteSecret(ctx context.Context, key string) error {
	delete(m.secretEntries, key)
	return nil
}
func (m *mockMCPStore) ListSecret(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	for k := range m.secretEntries {
		keys = append(keys, k)
	}
	return keys, nil
}
func (m *mockMCPStore) GetToken(ctx context.Context, tokenHash string) (*storage.TokenMeta, error) {
	if t, ok := m.tokens[tokenHash]; ok {
		return t, nil
	}
	return nil, storage.ErrNotFound
}
func (m *mockMCPStore) SetToken(ctx context.Context, tokenHash string, meta *storage.TokenMeta) error {
	m.tokens[tokenHash] = meta
	return nil
}
func (m *mockMCPStore) DeleteToken(ctx context.Context, tokenHash string) error {
	delete(m.tokens, tokenHash)
	return nil
}
func (m *mockMCPStore) ListTokens(ctx context.Context) ([]string, error) {
	var hashes []string
	for h := range m.tokens {
		hashes = append(hashes, h)
	}
	return hashes, nil
}
func (m *mockMCPStore) Close() error { return nil }

func TestNewServer(t *testing.T) {
	store := newMockMCPStore()
	server := NewServer(store, "1.0.0")

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
}

func TestKVGet(t *testing.T) {
	store := newMockMCPStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "test-key", Value: "test-value"})

	server := NewServer(store, "1.0.0")
	args := KVGetArgs{Key: "test-key"}

	result, _, err := server.kvGet(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("kvGet() error = %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("kvGet() returned no content")
	}
}

func TestKVGet_NotFound(t *testing.T) {
	store := newMockMCPStore()
	server := NewServer(store, "1.0.0")

	args := KVGetArgs{Key: "nonexistent"}
	result, _, err := server.kvGet(context.Background(), nil, args)

	if err != nil {
		t.Fatalf("kvGet() error = %v", err)
	}
	if result == nil {
		t.Fatal("kvGet() should return result for not found")
	}
}

func TestKVSet(t *testing.T) {
	store := newMockMCPStore()
	server := NewServer(store, "1.0.0")

	args := KVSetArgs{Key: "new-key", Value: "new-value"}
	result, _, err := server.kvSet(context.Background(), nil, args)

	if err != nil {
		t.Fatalf("kvSet() error = %v", err)
	}
	if result == nil {
		t.Fatal("kvSet() should return result")
	}

	entry, _ := store.GetKV(context.Background(), "new-key")
	if entry == nil {
		t.Error("kvSet() did not store the entry")
	}
}

func TestKVDelete(t *testing.T) {
	store := newMockMCPStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "delete-key", Value: "value"})

	server := NewServer(store, "1.0.0")
	args := KVDeleteArgs{Key: "delete-key"}

	result, _, err := server.kvDelete(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("kvDelete() error = %v", err)
	}
	if result == nil {
		t.Fatal("kvDelete() should return result")
	}
}

func TestKVList(t *testing.T) {
	store := newMockMCPStore()
	store.SetKV(context.Background(), &storage.Entry{Key: "key1", Value: "v1"})
	store.SetKV(context.Background(), &storage.Entry{Key: "key2", Value: "v2"})

	server := NewServer(store, "1.0.0")
	args := KVListArgs{}

	result, _, err := server.kvList(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("kvList() error = %v", err)
	}
	if result == nil {
		t.Fatal("kvList() should return result")
	}
}

func TestSecretGet(t *testing.T) {
	store := newMockMCPStore()
	store.SetSecret(context.Background(), &storage.SecretEntry{
		Entry: storage.Entry{Key: "db-password", Value: "secret123"},
		Type:  storage.SecretTypePassword,
	})

	server := NewServer(store, "1.0.0")
	args := SecretGetArgs{Key: "db-password"}

	result, _, err := server.secretGet(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("secretGet() error = %v", err)
	}
	if result == nil {
		t.Fatal("secretGet() should return result")
	}
}

func TestSecretSet(t *testing.T) {
	store := newMockMCPStore()
	server := NewServer(store, "1.0.0")

	args := SecretSetArgs{
		Key:   "api-key",
		Value: "sk-123",
		Type:  "api_key",
	}
	result, _, err := server.secretSet(context.Background(), nil, args)

	if err != nil {
		t.Fatalf("secretSet() error = %v", err)
	}
	if result == nil {
		t.Fatal("secretSet() should return result")
	}

	secret, _ := store.GetSecret(context.Background(), "api-key")
	if secret == nil {
		t.Error("secretSet() did not store the entry")
	}
}

func TestSecretSet_InvalidType(t *testing.T) {
	store := newMockMCPStore()
	server := NewServer(store, "1.0.0")

	args := SecretSetArgs{
		Key:   "test",
		Value: "value",
		Type:  "invalid_type",
	}

	_, _, err := server.secretSet(context.Background(), nil, args)
	if err == nil {
		t.Error("secretSet() should return error for invalid type")
	}
}

func TestSecretList(t *testing.T) {
	store := newMockMCPStore()
	store.SetSecret(context.Background(), &storage.SecretEntry{
		Entry: storage.Entry{Key: "secret1", Value: "v1"},
		Type:  storage.SecretTypePassword,
	})

	server := NewServer(store, "1.0.0")
	args := SecretListArgs{}

	result, _, err := server.secretList(context.Background(), nil, args)
	if err != nil {
		t.Fatalf("secretList() error = %v", err)
	}
	if result == nil {
		t.Fatal("secretList() should return result")
	}
}
