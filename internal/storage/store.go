// Package storage provides the storage interface and implementations.
package storage

import (
	"context"
	"errors"
	"time"

	"github.com/skys-mission/keysdk"
)

// ErrNotFound is returned when an entry is not found.
var ErrNotFound = errors.New("entry not found")

// SecretType is an alias to keysdk.SecretType for storage layer compatibility.
type SecretType = keysdk.SecretType

// Secret type constants - aliases to keysdk constants.
const (
	SecretTypePassword    SecretType = keysdk.SecretTypePassword
	SecretTypeAPIKey      SecretType = keysdk.SecretTypeAPIKey
	SecretTypeCertificate SecretType = keysdk.SecretTypeCertificate
	SecretTypePrivateKey  SecretType = keysdk.SecretTypePrivateKey
	SecretTypeToken       SecretType = keysdk.SecretTypeToken
	SecretTypeOther       SecretType = keysdk.SecretTypeOther
)

// Metadata holds additional information for a KV entry.
type Metadata struct {
	Description string            `json:"description,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
}

// Entry represents a KV storage entry.
type Entry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Metadata  Metadata  `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int64     `json:"version"`
}

// SecretEntry represents a secret storage entry.
type SecretEntry struct {
	Entry
	Type SecretType `json:"type"`
}

// TokenMeta holds metadata for an access token.
type TokenMeta struct {
	Name      string     `json:"name"`
	Type      string     `json:"type"` // client or mcp
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// Store defines the storage interface.
type Store interface {
	// KV operations
	GetKV(ctx context.Context, key string) (*Entry, error)
	SetKV(ctx context.Context, entry *Entry) error
	DeleteKV(ctx context.Context, key string) error
	ListKV(ctx context.Context, prefix string) ([]string, error)

	// Secret operations
	GetSecret(ctx context.Context, key string) (*SecretEntry, error)
	SetSecret(ctx context.Context, entry *SecretEntry) error
	DeleteSecret(ctx context.Context, key string) error
	ListSecret(ctx context.Context, prefix string) ([]string, error)

	// Token operations
	GetToken(ctx context.Context, tokenHash string) (*TokenMeta, error)
	SetToken(ctx context.Context, tokenHash string, meta *TokenMeta) error
	DeleteToken(ctx context.Context, tokenHash string) error
	ListTokens(ctx context.Context) ([]string, error)

	// Lifecycle
	Close() error
}
