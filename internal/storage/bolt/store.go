// Package bolt provides BoltDB-based storage implementation.
package bolt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/skys-mission/key-agent/internal/storage"
	"github.com/skys-mission/key-agent/internal/storage/crypto"
	"go.etcd.io/bbolt"
)

// Bucket names
var (
	bucketKV     = []byte("kv")
	bucketSecret = []byte("secrets")
	bucketToken  = []byte("tokens")
	bucketMeta   = []byte("meta")
)

// Store implements storage.Store using BoltDB with encryption.
type Store struct {
	db  *bbolt.DB
	enc *crypto.Encryptor
}

// New creates a new BoltDB store with encryption.
func New(dbPath string, encryptor *crypto.Encryptor) (*Store, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create buckets
	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := [][]byte{bucketKV, bucketSecret, bucketToken, bucketMeta}
		for _, name := range buckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", name, err)
			}
		}
		return nil
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{
		db:  db,
		enc: encryptor,
	}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// GetKV retrieves a KV entry by key.
func (s *Store) GetKV(ctx context.Context, key string) (*storage.Entry, error) {
	var entry storage.Entry

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketKV)
		if b == nil {
			return errors.New("kv bucket not found")
		}

		encrypted := b.Get([]byte(key))
		if encrypted == nil {
			return storage.ErrNotFound
		}

		// Decrypt
		plaintext, err := s.enc.Decrypt(encrypted)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		// Unmarshal
		if err := json.Unmarshal(plaintext, &entry); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// SetKV stores a KV entry.
func (s *Store) SetKV(ctx context.Context, entry *storage.Entry) error {
	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now

	// Marshal
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Encrypt
	encrypted, err := s.enc.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketKV)
		if b == nil {
			return errors.New("kv bucket not found")
		}

		return b.Put([]byte(entry.Key), encrypted)
	})
}

// DeleteKV deletes a KV entry.
func (s *Store) DeleteKV(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketKV)
		if b == nil {
			return errors.New("kv bucket not found")
		}

		if b.Get([]byte(key)) == nil {
			return storage.ErrNotFound
		}

		return b.Delete([]byte(key))
	})
}

// ListKV lists all KV keys with optional prefix filter.
func (s *Store) ListKV(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketKV)
		if b == nil {
			return errors.New("kv bucket not found")
		}

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			key := string(k)
			if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keys = append(keys, key)
			}
		}
		return nil
	})

	return keys, err
}

// GetSecret retrieves a secret entry by key.
func (s *Store) GetSecret(ctx context.Context, key string) (*storage.SecretEntry, error) {
	var entry storage.SecretEntry

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketSecret)
		if b == nil {
			return errors.New("secrets bucket not found")
		}

		encrypted := b.Get([]byte(key))
		if encrypted == nil {
			return storage.ErrNotFound
		}

		// Decrypt
		plaintext, err := s.enc.Decrypt(encrypted)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		// Unmarshal
		if err := json.Unmarshal(plaintext, &entry); err != nil {
			return fmt.Errorf("unmarshal failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// SetSecret stores a secret entry.
func (s *Store) SetSecret(ctx context.Context, entry *storage.SecretEntry) error {
	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now

	// Marshal
	plaintext, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// Encrypt
	encrypted, err := s.enc.Encrypt(plaintext)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketSecret)
		if b == nil {
			return errors.New("secrets bucket not found")
		}

		return b.Put([]byte(entry.Key), encrypted)
	})
}

// DeleteSecret deletes a secret entry.
func (s *Store) DeleteSecret(ctx context.Context, key string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketSecret)
		if b == nil {
			return errors.New("secrets bucket not found")
		}

		if b.Get([]byte(key)) == nil {
			return storage.ErrNotFound
		}

		return b.Delete([]byte(key))
	})
}

// ListSecret lists all secret keys with optional prefix filter.
func (s *Store) ListSecret(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketSecret)
		if b == nil {
			return errors.New("secrets bucket not found")
		}

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			key := string(k)
			if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
				keys = append(keys, key)
			}
		}
		return nil
	})

	return keys, err
}

// GetToken retrieves token metadata by hash.
func (s *Store) GetToken(ctx context.Context, tokenHash string) (*storage.TokenMeta, error) {
	var meta storage.TokenMeta

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketToken)
		if b == nil {
			return errors.New("tokens bucket not found")
		}

		data := b.Get([]byte(tokenHash))
		if data == nil {
			return storage.ErrNotFound
		}

		return json.Unmarshal(data, &meta)
	})

	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// SetToken stores token metadata.
func (s *Store) SetToken(ctx context.Context, tokenHash string, meta *storage.TokenMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketToken)
		if b == nil {
			return errors.New("tokens bucket not found")
		}

		return b.Put([]byte(tokenHash), data)
	})
}

// DeleteToken deletes a token.
func (s *Store) DeleteToken(ctx context.Context, tokenHash string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketToken)
		if b == nil {
			return errors.New("tokens bucket not found")
		}

		return b.Delete([]byte(tokenHash))
	})
}

// ListTokens lists all token hashes.
func (s *Store) ListTokens(ctx context.Context) ([]string, error) {
	var hashes []string

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketToken)
		if b == nil {
			return errors.New("tokens bucket not found")
		}

		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			hashes = append(hashes, string(k))
		}
		return nil
	})

	return hashes, err
}
