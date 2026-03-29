// Package masterkey provides the OS keyring backend.
package masterkey

import (
	"encoding/base64"

	keyring "github.com/zalando/go-keyring"
)

const (
	keyringService = "key-agent"
	keyringUser    = "master-key"
)

// KeyringBackend stores the master key in the OS keyring.
type KeyringBackend struct {
	available bool
}

// NewKeyringBackend creates a new keyring backend.
func NewKeyringBackend() *KeyringBackend {
	return &KeyringBackend{
		available: true, // Will be checked on first use
	}
}

// Name returns the backend name.
func (b *KeyringBackend) Name() string {
	return "keyring"
}

// Available returns true if the OS keyring is accessible.
func (b *KeyringBackend) Available() bool {
	// Try to access the keyring to check availability
	// This is a lightweight check
	_, err := keyring.Get(keyringService, keyringUser+"_check")
	if err != nil && err != keyring.ErrNotFound {
		// Keyring is not accessible
		b.available = false
		return false
	}
	return true
}

// Get retrieves the master key from the keyring.
func (b *KeyringBackend) Get() ([]byte, error) {
	encoded, err := keyring.Get(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// Set stores the master key in the keyring.
func (b *KeyringBackend) Set(key []byte) error {
	encoded := base64.StdEncoding.EncodeToString(key)
	return keyring.Set(keyringService, keyringUser, encoded)
}

// Delete removes the master key from the keyring.
func (b *KeyringBackend) Delete() error {
	err := keyring.Delete(keyringService, keyringUser)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}
