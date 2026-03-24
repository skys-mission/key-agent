// Package masterkey provides the encrypted file backend.
package masterkey

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skys-mission/key-agent/internal/storage/crypto"
	"golang.org/x/crypto/scrypt"
)

const (
	// Scrypt parameters for key derivation
	scryptN = 32768 // CPU cost
	scryptR = 8     // Memory cost
	scryptP = 1     // Parallelism
	keyLen  = 32    // Derived key length

	// File names
	keyFileName = "master.key"
)

// FileBackend stores the master key in an encrypted file.
// This is the fallback backend when no hardware security is available.
type FileBackend struct {
	dataDir string
}

// NewFileBackend creates a new file backend.
func NewFileBackend() *FileBackend {
	homeDir, _ := os.UserHomeDir()
	return &FileBackend{
		dataDir: filepath.Join(homeDir, ".key-agent", "data"),
	}
}

// NewFileBackendWithDir creates a new file backend with a custom data directory.
func NewFileBackendWithDir(dataDir string) *FileBackend {
	return &FileBackend{
		dataDir: dataDir,
	}
}

// Name returns the backend name.
func (b *FileBackend) Name() string {
	return "file"
}

// Available always returns true for file backend.
func (b *FileBackend) Available() bool {
	return true
}

// Get retrieves the master key from the encrypted file.
// This requires the user's passphrase.
func (b *FileBackend) Get() ([]byte, error) {
	keyPath := filepath.Join(b.dataDir, keyFileName)

	data, err := os.ReadFile(keyPath)
	if os.IsNotExist(err) {
		return nil, ErrKeyNotFound
	}
	if err != nil {
		return nil, err
	}

	var encryptedKey encryptedKeyFile
	if err := json.Unmarshal(data, &encryptedKey); err != nil {
		return nil, fmt.Errorf("invalid key file format: %w", err)
	}

	// Need passphrase to decrypt
	passphrase, err := b.getPassphrase("Enter passphrase to unlock master key: ")
	if err != nil {
		return nil, err
	}

	key, err := decryptMasterKey(&encryptedKey, passphrase)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return key, nil
}

// Set stores the master key in an encrypted file.
// This will prompt for a passphrase.
func (b *FileBackend) Set(key []byte) error {
	if err := os.MkdirAll(b.dataDir, 0700); err != nil {
		return err
	}

	// Prompt for passphrase
	passphrase, err := b.getPassphrase("Enter new passphrase for master key: ")
	if err != nil {
		return err
	}

	// Confirm passphrase
	confirm, err := b.getPassphrase("Confirm passphrase: ")
	if err != nil {
		return err
	}

	if passphrase != confirm {
		return errors.New("passphrases do not match")
	}

	encryptedKey, err := encryptMasterKey(key, passphrase)
	if err != nil {
		return err
	}

	data, err := json.Marshal(encryptedKey)
	if err != nil {
		return err
	}

	keyPath := filepath.Join(b.dataDir, keyFileName)
	return os.WriteFile(keyPath, data, 0600)
}

// Delete removes the master key file.
func (b *FileBackend) Delete() error {
	keyPath := filepath.Join(b.dataDir, keyFileName)
	err := os.Remove(keyPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// getPassphrase prompts the user for a passphrase.
func (b *FileBackend) getPassphrase(prompt string) (string, error) {
	fmt.Print(prompt)
	var passphrase string
	_, err := fmt.Scanln(&passphrase)
	if err != nil {
		return "", err
	}
	return passphrase, nil
}

// encryptedKeyFile holds the encrypted master key and its derivation parameters.
type encryptedKeyFile struct {
	Salt       string `json:"salt"`        // Base64 encoded salt
	Nonce      string `json:"nonce"`       // Base64 encoded nonce
	Ciphertext string `json:"ciphertext"`  // Base64 encoded encrypted key
}

// encryptMasterKey encrypts the master key with a passphrase.
func encryptMasterKey(key []byte, passphrase string) (*encryptedKeyFile, error) {
	// Generate random salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Derive encryption key from passphrase
	encKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return nil, err
	}

	// Create encryptor
	enc, err := crypto.NewEncryptor(encKey)
	if err != nil {
		return nil, err
	}
	defer enc.Destroy()

	// Encrypt the master key
	ciphertext, err := enc.Encrypt(key)
	if err != nil {
		return nil, err
	}

	// Extract nonce (first 12 bytes of ciphertext)
	nonce := ciphertext[:crypto.NonceSize]
	ciphertext = ciphertext[crypto.NonceSize:]

	return &encryptedKeyFile{
		Salt:       base64.StdEncoding.EncodeToString(salt),
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// decryptMasterKey decrypts the master key with a passphrase.
func decryptMasterKey(encryptedKey *encryptedKeyFile, passphrase string) ([]byte, error) {
	// Decode base64
	salt, err := base64.StdEncoding.DecodeString(encryptedKey.Salt)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(encryptedKey.Nonce)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey.Ciphertext)
	if err != nil {
		return nil, err
	}

	// Derive encryption key from passphrase
	encKey, err := scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, keyLen)
	if err != nil {
		return nil, err
	}

	// Create encryptor
	enc, err := crypto.NewEncryptor(encKey)
	if err != nil {
		return nil, err
	}
	defer enc.Destroy()

	// Reconstruct full ciphertext with nonce
	fullCiphertext := make([]byte, crypto.NonceSize+len(ciphertext))
	copy(fullCiphertext[:crypto.NonceSize], nonce)
	copy(fullCiphertext[crypto.NonceSize:], ciphertext)

	// Decrypt
	return enc.Decrypt(fullCiphertext)
}
