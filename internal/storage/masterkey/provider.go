// Package masterkey provides master key management with multiple backends.
package masterkey

import (
	"crypto/rand"
	"errors"
	"fmt"
)

// Master key identifier stored in the database.
const MasterKeyID = "master-key"

// Backend defines the interface for master key storage backends.
type Backend interface {
	// Name returns the backend name.
	Name() string

	// Available returns true if this backend is usable on the current system.
	Available() bool

	// Get retrieves the master key.
	Get() ([]byte, error)

	// Set stores the master key.
	Set(key []byte) error

	// Delete removes the master key.
	Delete() error
}

var (
	// ErrNotAvailable is returned when the backend is not available.
	ErrNotAvailable = errors.New("backend not available")

	// ErrKeyNotFound is returned when the master key is not found.
	ErrKeyNotFound = errors.New("master key not found")
)

// Provider manages master key storage with automatic backend selection.
type Provider struct {
	backends []Backend
}

// NewProvider creates a new master key provider with the given backends.
// Backends are tried in order until one succeeds.
func NewProvider(backends ...Backend) *Provider {
	return &Provider{
		backends: backends,
	}
}

// DefaultProvider creates a provider with default backend priority:
// 1. OS Keyring (for desktop environments)
// 2. TPM 2.0 (for servers with hardware support)
// 3. Encrypted file (fallback, requires passphrase)
func DefaultProvider() *Provider {
	return NewProvider(
		NewKeyringBackend(),
		NewTPMBackend(),
		NewFileBackend(),
	)
}

// GetOrCreate retrieves the master key, creating a new one if it doesn't exist.
func (p *Provider) GetOrCreate() ([]byte, error) {
	// Try to get existing key
	key, err := p.Get()
	if err == nil {
		return key, nil
	}

	if !errors.Is(err, ErrKeyNotFound) {
		return nil, fmt.Errorf("failed to get master key: %w", err)
	}

	// Generate new key
	key, err = GenerateMasterKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key: %w", err)
	}

	// Store the new key
	if err := p.Set(key); err != nil {
		return nil, fmt.Errorf("failed to store master key: %w", err)
	}

	return key, nil
}

// Get retrieves the master key from the first available backend.
func (p *Provider) Get() ([]byte, error) {
	var lastErr error

	for _, backend := range p.backends {
		if !backend.Available() {
			continue
		}

		key, err := backend.Get()
		if err == nil {
			return key, nil
		}

		if errors.Is(err, ErrKeyNotFound) {
			// Key not found in this backend, try next
			lastErr = err
			continue
		}

		// Other error, record and try next
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, ErrKeyNotFound
}

// Set stores the master key in all available backends.
func (p *Provider) Set(key []byte) error {
	var stored bool

	for _, backend := range p.backends {
		if !backend.Available() {
			continue
		}

		if err := backend.Set(key); err != nil {
			continue
		}

		stored = true
		break // Success, no need to try other backends
	}

	if !stored {
		return errors.New("no available backend to store master key")
	}

	return nil
}

// GenerateMasterKey generates a new 32-byte master key.
func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
