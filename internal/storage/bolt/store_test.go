package bolt

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/skys-mission/key-agent/internal/storage"
	"github.com/skys-mission/key-agent/internal/storage/crypto"
)

func newTestStore(t *testing.T) (*Store, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "key-agent-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	key := make([]byte, crypto.KeySize)
	enc, err := crypto.NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := New(dbPath, enc)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestKVOperations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	// Test SetKV and GetKV
	t.Run("set and get", func(t *testing.T) {
		entry := &storage.Entry{
			Key:   "test-key",
			Value: "test-value",
		}

		if err := store.SetKV(ctx, entry); err != nil {
			t.Fatalf("SetKV() error: %v", err)
		}

		got, err := store.GetKV(ctx, "test-key")
		if err != nil {
			t.Fatalf("GetKV() error: %v", err)
		}

		if got.Key != entry.Key {
			t.Errorf("GetKV() key = %v, want %v", got.Key, entry.Key)
		}
		if got.Value != entry.Value {
			t.Errorf("GetKV() value = %v, want %v", got.Value, entry.Value)
		}
		if got.CreatedAt.IsZero() {
			t.Error("GetKV() CreatedAt should be set")
		}
		if got.UpdatedAt.IsZero() {
			t.Error("GetKV() UpdatedAt should be set")
		}
	})

	// Test GetKV not found
	t.Run("get not found", func(t *testing.T) {
		_, err := store.GetKV(ctx, "non-existent")
		if err != storage.ErrNotFound {
			t.Errorf("GetKV() error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	// Test DeleteKV
	t.Run("delete", func(t *testing.T) {
		entry := &storage.Entry{Key: "delete-key", Value: "value"}
		store.SetKV(ctx, entry)

		if err := store.DeleteKV(ctx, "delete-key"); err != nil {
			t.Fatalf("DeleteKV() error: %v", err)
		}

		_, err := store.GetKV(ctx, "delete-key")
		if err != storage.ErrNotFound {
			t.Errorf("GetKV() after delete error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	// Test DeleteKV not found
	t.Run("delete not found", func(t *testing.T) {
		err := store.DeleteKV(ctx, "non-existent")
		if err != storage.ErrNotFound {
			t.Errorf("DeleteKV() error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	// Test ListKV
	t.Run("list", func(t *testing.T) {
		// Clean up existing keys
		keys, _ := store.ListKV(ctx, "")
		for _, k := range keys {
			store.DeleteKV(ctx, k)
		}

		// Add test keys
		testKeys := []string{"a/b/c", "a/b/d", "a/e", "x/y"}
		for _, k := range testKeys {
			store.SetKV(ctx, &storage.Entry{Key: k, Value: "v"})
		}

		// List all
		all, err := store.ListKV(ctx, "")
		if err != nil {
			t.Fatalf("ListKV() error: %v", err)
		}
		if len(all) != len(testKeys) {
			t.Errorf("ListKV() count = %d, want %d", len(all), len(testKeys))
		}

		// List with prefix
		prefix, err := store.ListKV(ctx, "a/b")
		if err != nil {
			t.Fatalf("ListKV() with prefix error: %v", err)
		}
		if len(prefix) != 2 {
			t.Errorf("ListKV() with prefix count = %d, want 2", len(prefix))
		}
	})

	// Test update preserves created_at
	t.Run("update preserves created_at", func(t *testing.T) {
		entry := &storage.Entry{Key: "update-key", Value: "v1"}
		store.SetKV(ctx, entry)

		original, _ := store.GetKV(ctx, "update-key")
		originalCreatedAt := original.CreatedAt

		time.Sleep(10 * time.Millisecond)

		updated := &storage.Entry{Key: "update-key", Value: "v2"}
		updated.CreatedAt = originalCreatedAt
		store.SetKV(ctx, updated)

		got, _ := store.GetKV(ctx, "update-key")
		if got.CreatedAt != originalCreatedAt {
			t.Error("Update should preserve CreatedAt")
		}
		if got.Value != "v2" {
			t.Errorf("Update value = %v, want v2", got.Value)
		}
	})
}

func TestSecretOperations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("set and get secret", func(t *testing.T) {
		entry := &storage.SecretEntry{
			Entry: storage.Entry{
				Key:   "db-password",
				Value: "secret123",
			},
			Type: storage.SecretTypePassword,
		}

		if err := store.SetSecret(ctx, entry); err != nil {
			t.Fatalf("SetSecret() error: %v", err)
		}

		got, err := store.GetSecret(ctx, "db-password")
		if err != nil {
			t.Fatalf("GetSecret() error: %v", err)
		}

		if got.Key != entry.Key {
			t.Errorf("GetSecret() key = %v, want %v", got.Key, entry.Key)
		}
		if got.Value != entry.Value {
			t.Errorf("GetSecret() value = %v, want %v", got.Value, entry.Value)
		}
		if got.Type != storage.SecretTypePassword {
			t.Errorf("GetSecret() type = %v, want %v", got.Type, storage.SecretTypePassword)
		}
	})

	t.Run("get not found", func(t *testing.T) {
		_, err := store.GetSecret(ctx, "non-existent")
		if err != storage.ErrNotFound {
			t.Errorf("GetSecret() error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	t.Run("delete secret", func(t *testing.T) {
		entry := &storage.SecretEntry{
			Entry: storage.Entry{Key: "delete-secret", Value: "v"},
			Type:  storage.SecretTypeAPIKey,
		}
		store.SetSecret(ctx, entry)

		if err := store.DeleteSecret(ctx, "delete-secret"); err != nil {
			t.Fatalf("DeleteSecret() error: %v", err)
		}

		_, err := store.GetSecret(ctx, "delete-secret")
		if err != storage.ErrNotFound {
			t.Errorf("GetSecret() after delete error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	t.Run("list secrets", func(t *testing.T) {
		// Clean up
		keys, _ := store.ListSecret(ctx, "")
		for _, k := range keys {
			store.DeleteSecret(ctx, k)
		}

		// Add test secrets
		testKeys := []string{"secret1", "secret2", "prefix/secret"}
		for _, k := range testKeys {
			store.SetSecret(ctx, &storage.SecretEntry{
				Entry: storage.Entry{Key: k, Value: "v"},
				Type:  storage.SecretTypeOther,
			})
		}

		all, err := store.ListSecret(ctx, "")
		if err != nil {
			t.Fatalf("ListSecret() error: %v", err)
		}
		if len(all) != len(testKeys) {
			t.Errorf("ListSecret() count = %d, want %d", len(all), len(testKeys))
		}
	})
}

func TestTokenOperations(t *testing.T) {
	store, cleanup := newTestStore(t)
	defer cleanup()
	ctx := context.Background()

	t.Run("set and get token", func(t *testing.T) {
		now := time.Now()
		meta := &storage.TokenMeta{
			Name:      "test-token",
			Type:      "client",
			CreatedAt: now,
		}

		tokenHash := "hash123"

		if err := store.SetToken(ctx, tokenHash, meta); err != nil {
			t.Fatalf("SetToken() error: %v", err)
		}

		got, err := store.GetToken(ctx, tokenHash)
		if err != nil {
			t.Fatalf("GetToken() error: %v", err)
		}

		if got.Name != meta.Name {
			t.Errorf("GetToken() name = %v, want %v", got.Name, meta.Name)
		}
		if got.Type != meta.Type {
			t.Errorf("GetToken() type = %v, want %v", got.Type, meta.Type)
		}
	})

	t.Run("get token not found", func(t *testing.T) {
		_, err := store.GetToken(ctx, "non-existent-hash")
		if err != storage.ErrNotFound {
			t.Errorf("GetToken() error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	t.Run("delete token", func(t *testing.T) {
		hash := "delete-hash"
		meta := &storage.TokenMeta{Name: "delete-me", Type: "mcp", CreatedAt: time.Now()}
		store.SetToken(ctx, hash, meta)

		if err := store.DeleteToken(ctx, hash); err != nil {
			t.Fatalf("DeleteToken() error: %v", err)
		}

		_, err := store.GetToken(ctx, hash)
		if err != storage.ErrNotFound {
			t.Errorf("GetToken() after delete error = %v, want %v", err, storage.ErrNotFound)
		}
	})

	t.Run("list tokens", func(t *testing.T) {
		// Clean up
		hashes, _ := store.ListTokens(ctx)
		for _, h := range hashes {
			store.DeleteToken(ctx, h)
		}

		// Add test tokens
		testHashes := []string{"hash1", "hash2", "hash3"}
		for _, h := range testHashes {
			store.SetToken(ctx, h, &storage.TokenMeta{
				Name:      "token-" + h,
				Type:      "client",
				CreatedAt: time.Now(),
			})
		}

		all, err := store.ListTokens(ctx)
		if err != nil {
			t.Fatalf("ListTokens() error: %v", err)
		}
		if len(all) != len(testHashes) {
			t.Errorf("ListTokens() count = %d, want %d", len(all), len(testHashes))
		}
	})
}

func TestClose(t *testing.T) {
	store, cleanup := newTestStore(t)
	cleanup() // Should not panic

	// Second close should be safe
	if err := store.Close(); err == nil {
		t.Log("Double close handled gracefully")
	}
}
