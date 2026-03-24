// Package crypto provides encryption utilities using AES-GCM.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const (
	// KeySize is the size of the encryption key in bytes (AES-256).
	KeySize = 32

	// NonceSize is the size of the GCM nonce in bytes.
	NonceSize = 12

	// Overhead is the additional size added by AES-GCM authentication tag.
	Overhead = 16
)

var (
	// ErrInvalidKeySize is returned when the key size is not 32 bytes.
	ErrInvalidKeySize = errors.New("key must be 32 bytes")

	// ErrInvalidCiphertext is returned when the ciphertext is too short.
	ErrInvalidCiphertext = errors.New("ciphertext too short")

	// ErrDecryptFailed is returned when decryption fails.
	ErrDecryptFailed = errors.New("decryption failed")
)

// Encryptor provides AES-256-GCM encryption and decryption.
type Encryptor struct {
	key   []byte
	aead  cipher.AEAD
}

// NewEncryptor creates a new Encryptor with the given key.
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Store key in a secure manner (copy to prevent external modification)
	keyCopy := make([]byte, KeySize)
	copy(keyCopy, key)

	return &Encryptor{
		key:  keyCopy,
		aead: aead,
	}, nil
}

// Encrypt encrypts plaintext using AES-GCM.
// The output format is: nonce (12 bytes) || ciphertext || tag (16 bytes).
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := e.aead.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext
	result := make([]byte, NonceSize+len(ciphertext))
	copy(result[:NonceSize], nonce)
	copy(result[NonceSize:], ciphertext)

	return result, nil
}

// Decrypt decrypts ciphertext using AES-GCM.
// The input format is: nonce (12 bytes) || ciphertext || tag.
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < NonceSize+Overhead {
		return nil, ErrInvalidCiphertext
	}

	nonce := ciphertext[:NonceSize]
	ciphertext = ciphertext[NonceSize:]

	plaintext, err := e.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	return plaintext, nil
}

// Destroy securely clears the encryption key from memory.
func (e *Encryptor) Destroy() {
	// Overwrite key with zeros
	for i := range e.key {
		e.key[i] = 0
	}
	e.key = nil
	e.aead = nil
}

// GenerateKey generates a new random encryption key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}
