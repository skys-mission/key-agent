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

// Environment variable names for master key configuration.
const (
	// EnvMasterKey is the environment variable for directly providing a master key.
	// The value should be a base64-encoded 32-byte key.
	// When set, it takes precedence over file-based storage.
	EnvMasterKey = "KEY_AGENT_MASTER_KEY"

	// EnvPassphrase is the environment variable for the passphrase used to
	// encrypt/decrypt the master key file.
	EnvPassphrase = "KEY_AGENT_PASSPHRASE"
)

const (
	// Scrypt parameters for key derivation
	scryptN = 32768 // CPU cost
	scryptR = 8     // Memory cost
	scryptP = 1     // Parallelism
	keyLen  = 32    // Derived key length

	// File names
	keyFileName        = "master.key"
	passphraseFileName = ".passphrase" // Hidden file for auto-generated passphrase
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

// Available returns true if either:
// - KEY_AGENT_MASTER_KEY environment variable is set, or
// - The backend can access the file system (always true for file backend).
func (b *FileBackend) Available() bool {
	// Environment variable takes precedence
	if os.Getenv(EnvMasterKey) != "" {
		return true
	}
	return true
}

// Get retrieves the master key.
// Priority: KEY_AGENT_MASTER_KEY env > encrypted file.
func (b *FileBackend) Get() ([]byte, error) {
	// Check environment variable first (direct master key injection)
	if key, err := b.getFromEnv(); err == nil {
		return key, nil
	}

	// Fall back to file-based storage
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
// If KEY_AGENT_MASTER_KEY environment variable is set, this is a no-op
// (the key is already provided via environment).
// Otherwise, this will prompt for a passphrase (or auto-generate in non-interactive mode).
func (b *FileBackend) Set(key []byte) error {
	// If master key is provided via environment, skip file storage
	if os.Getenv(EnvMasterKey) != "" {
		return nil
	}

	if err := os.MkdirAll(b.dataDir, 0700); err != nil {
		return err
	}

	// Get passphrase (will auto-generate in non-interactive mode)
	passphrase, err := b.getPassphrase("Enter new passphrase for master key: ")
	if err != nil {
		return err
	}

	// Confirm passphrase only in interactive mode
	if b.hasTTY() {
		confirm, err := b.getPassphrase("Confirm passphrase: ")
		if err != nil {
			return err
		}
		if passphrase != confirm {
			return errors.New("passphrases do not match")
		}
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
// If KEY_AGENT_PASSPHRASE environment variable is set, it uses that value.
// If running in a container without TTY, it auto-generates and saves a passphrase.
func (b *FileBackend) getPassphrase(prompt string) (string, error) {
	// Check environment variable first
	if passphrase := os.Getenv(EnvPassphrase); passphrase != "" {
		return passphrase, nil
	}

	// Try to read saved passphrase file (for container auto-generated passphrase)
	savedPassphrase, err := b.readSavedPassphrase()
	if err == nil && savedPassphrase != "" {
		return savedPassphrase, nil
	}

	// Check if running in container without TTY (non-interactive)
	if !b.hasTTY() {
		// Auto-generate passphrase for container environment
		return b.generateAndSavePassphrase()
	}

	// Interactive prompt
	fmt.Print(prompt)
	var passphrase string
	_, err = fmt.Scanln(&passphrase)
	if err != nil {
		return "", err
	}
	return passphrase, nil
}

// getFromEnv retrieves the master key from KEY_AGENT_MASTER_KEY environment variable.
// The value must be a base64-encoded 32-byte key.
// Returns ErrKeyNotFound if not set or invalid.
func (b *FileBackend) getFromEnv() ([]byte, error) {
	envKey := os.Getenv(EnvMasterKey)
	if envKey == "" {
		return nil, ErrKeyNotFound
	}

	key, err := base64.StdEncoding.DecodeString(envKey)
	if err != nil {
		return nil, fmt.Errorf("invalid master key encoding: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes, got %d", len(key))
	}

	return key, nil
}

// hasTTY checks if the process has an interactive terminal.
// Returns false in containers or non-interactive environments.
func (b *FileBackend) hasTTY() bool {
	// Check if running in a container
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return false
	}

	// Check stdin for TTY
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// Check if stdin is a character device (interactive terminal)
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// Also check if stdout is a TTY (for more reliable detection)
	fo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return (fo.Mode() & os.ModeCharDevice) != 0
}

// readSavedPassphrase reads the auto-generated passphrase from file.
func (b *FileBackend) readSavedPassphrase() (string, error) {
	path := filepath.Join(b.dataDir, passphraseFileName)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// generateAndSavePassphrase generates a random passphrase and saves it to file.
func (b *FileBackend) generateAndSavePassphrase() (string, error) {
	// Generate random passphrase
	passphrase := make([]byte, 32)
	if _, err := rand.Read(passphrase); err != nil {
		return "", fmt.Errorf("failed to generate passphrase: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(passphrase)

	// Ensure directory exists
	if err := os.MkdirAll(b.dataDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}

	// Save passphrase to file
	path := filepath.Join(b.dataDir, passphraseFileName)
	if err := os.WriteFile(path, []byte(encoded), 0600); err != nil {
		return "", fmt.Errorf("failed to save passphrase: %w", err)
	}

	return encoded, nil
}

// encryptedKeyFile holds the encrypted master key and its derivation parameters.
type encryptedKeyFile struct {
	Salt       string `json:"salt"`       // Base64 encoded salt
	Nonce      string `json:"nonce"`      // Base64 encoded nonce
	Ciphertext string `json:"ciphertext"` // Base64 encoded encrypted key
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
