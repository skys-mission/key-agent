package crypto

import (
	"bytes"
	"testing"
)

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr error
	}{
		{"valid 32-byte key", make([]byte, 32), nil},
		{"invalid 16-byte key", make([]byte, 16), ErrInvalidKeySize},
		{"invalid 64-byte key", make([]byte, 64), ErrInvalidKeySize},
		{"nil key", nil, ErrInvalidKeySize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := NewEncryptor(tt.key)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewEncryptor() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewEncryptor() unexpected error: %v", err)
				return
			}
			if enc == nil {
				t.Error("NewEncryptor() returned nil")
				return
			}
			enc.Destroy()
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}

	enc, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("NewEncryptor() error: %v", err)
	}
	defer enc.Destroy()

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"single byte", []byte{0x00}},
		{"short text", []byte("hello")},
		{"long text", bytes.Repeat([]byte("x"), 1000)},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xfe, 0xff}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Errorf("Encrypt() error: %v", err)
				return
			}

			// Ciphertext should be longer than plaintext (nonce + tag)
			minLen := len(tt.plaintext) + NonceSize + Overhead
			if len(ciphertext) < minLen {
				t.Errorf("Encrypt() ciphertext length = %d, want at least %d", len(ciphertext), minLen)
			}

			// Ciphertext should differ from plaintext
			if len(tt.plaintext) > 0 && bytes.Equal(tt.plaintext, ciphertext[NonceSize:len(tt.plaintext)+NonceSize]) {
				t.Error("Encrypt() ciphertext should not contain plaintext")
			}

			decrypted, err := enc.Decrypt(ciphertext)
			if err != nil {
				t.Errorf("Decrypt() error: %v", err)
				return
			}

			if !bytes.Equal(tt.plaintext, decrypted) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewEncryptor(key)
	defer enc.Destroy()

	plaintext := []byte("same plaintext")

	ct1, _ := enc.Encrypt(plaintext)
	ct2, _ := enc.Encrypt(plaintext)

	if bytes.Equal(ct1, ct2) {
		t.Error("Encrypt() should produce different ciphertext for same plaintext (random nonce)")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewEncryptor(key)
	defer enc.Destroy()

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    error
	}{
		{"empty", []byte{}, ErrInvalidCiphertext},
		{"too short", []byte{0x00, 0x01, 0x02}, ErrInvalidCiphertext},
		{"nonce only", make([]byte, NonceSize), ErrInvalidCiphertext},
		{"corrupted", append(make([]byte, NonceSize), make([]byte, Overhead+10)...), ErrDecryptFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.ciphertext)
			if err != tt.wantErr {
				t.Errorf("Decrypt() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	enc1, _ := NewEncryptor(key1)
	defer enc1.Destroy()

	enc2, _ := NewEncryptor(key2)
	defer enc2.Destroy()

	plaintext := []byte("secret data")
	ciphertext, _ := enc1.Encrypt(plaintext)

	_, err := enc2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypt() should fail with wrong key")
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error: %v", err)
	}
	if len(key1) != KeySize {
		t.Errorf("GenerateKey() length = %d, want %d", len(key1), KeySize)
	}

	key2, _ := GenerateKey()
	if bytes.Equal(key1, key2) {
		t.Error("GenerateKey() should produce different keys")
	}
}

func TestDestroy(t *testing.T) {
	key, _ := GenerateKey()
	enc, _ := NewEncryptor(key)

	enc.Destroy()

	// After destroy, key should be nil
	if enc.key != nil {
		t.Error("Destroy() should clear the key")
	}
	if enc.aead != nil {
		t.Error("Destroy() should clear the aead")
	}
}
