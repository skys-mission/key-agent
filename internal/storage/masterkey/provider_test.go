package masterkey

import (
	"errors"
	"testing"
)

// mockBackend implements Backend for testing
type mockBackend struct {
	available bool
	key       []byte
	getErr    error
	setErr    error
	delErr    error
}

func (m *mockBackend) Name() string         { return "mock" }
func (m *mockBackend) Available() bool      { return m.available }
func (m *mockBackend) Get() ([]byte, error) { return m.key, m.getErr }
func (m *mockBackend) Set(key []byte) error { m.key = key; return m.setErr }
func (m *mockBackend) Delete() error        { m.key = nil; return m.delErr }

func TestProvider_Get(t *testing.T) {
	testKey := []byte("test-master-key-32-bytes-long")

	tests := []struct {
		name     string
		backends []Backend
		wantKey  []byte
		wantErr  error
	}{
		{
			name:     "get from first available backend",
			backends: []Backend{&mockBackend{available: true, key: testKey}},
			wantKey:  testKey,
			wantErr:  nil,
		},
		{
			name: "skip unavailable backend, use second",
			backends: []Backend{
				&mockBackend{available: false},
				&mockBackend{available: true, key: testKey},
			},
			wantKey: testKey,
			wantErr: nil,
		},
		{
			name: "key not found in any backend",
			backends: []Backend{
				&mockBackend{available: true, getErr: ErrKeyNotFound},
				&mockBackend{available: true, getErr: ErrKeyNotFound},
			},
			wantKey: nil,
			wantErr: ErrKeyNotFound,
		},
		{
			name: "no available backends",
			backends: []Backend{
				&mockBackend{available: false},
				&mockBackend{available: false},
			},
			wantKey: nil,
			wantErr: ErrKeyNotFound,
		},
		{
			name:     "no backends",
			backends: []Backend{},
			wantKey:  nil,
			wantErr:  ErrKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.backends...)
			key, err := p.Get()

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantKey != nil && string(key) != string(tt.wantKey) {
				t.Errorf("Get() key = %v, want %v", key, tt.wantKey)
			}
		})
	}
}

func TestProvider_Set(t *testing.T) {
	testKey := []byte("test-master-key-32-bytes-long")

	tests := []struct {
		name     string
		backends []Backend
		wantErr  bool
	}{
		{
			name:     "set to first available backend",
			backends: []Backend{&mockBackend{available: true}},
			wantErr:  false,
		},
		{
			name: "skip unavailable backend, set to second",
			backends: []Backend{
				&mockBackend{available: false},
				&mockBackend{available: true},
			},
			wantErr: false,
		},
		{
			name: "no available backends",
			backends: []Backend{
				&mockBackend{available: false},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.backends...)
			err := p.Set(testKey)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_GetOrCreate(t *testing.T) {
	testKey := []byte("test-master-key-32-bytes-long")

	t.Run("get existing key", func(t *testing.T) {
		p := NewProvider(&mockBackend{available: true, key: testKey})
		key, err := p.GetOrCreate()
		if err != nil {
			t.Errorf("GetOrCreate() error = %v", err)
		}
		if string(key) != string(testKey) {
			t.Errorf("GetOrCreate() key = %v, want %v", key, testKey)
		}
	})

	t.Run("create new key when not found", func(t *testing.T) {
		backend := &mockBackend{available: true, getErr: ErrKeyNotFound}
		p := NewProvider(backend)
		key, err := p.GetOrCreate()

		if err != nil {
			t.Errorf("GetOrCreate() error = %v", err)
		}
		if len(key) != 32 {
			t.Errorf("GetOrCreate() key length = %d, want 32", len(key))
		}
		if backend.key == nil {
			t.Error("GetOrCreate() should store new key in backend")
		}
	})
}

func TestGenerateMasterKey(t *testing.T) {
	key1, err := GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey() error = %v", err)
	}
	if len(key1) != 32 {
		t.Errorf("GenerateMasterKey() length = %d, want 32", len(key1))
	}

	key2, _ := GenerateMasterKey()
	if string(key1) == string(key2) {
		t.Error("GenerateMasterKey() should generate different keys")
	}
}

func TestEncryptDecryptMasterKey(t *testing.T) {
	testKey := []byte("test-master-key-32-bytes-long!!")
	passphrase := "test-passphrase"

	encrypted, err := encryptMasterKey(testKey, passphrase)
	if err != nil {
		t.Fatalf("encryptMasterKey() error = %v", err)
	}

	if encrypted.Salt == "" {
		t.Error("encryptMasterKey() should set salt")
	}
	if encrypted.Nonce == "" {
		t.Error("encryptMasterKey() should set nonce")
	}
	if encrypted.Ciphertext == "" {
		t.Error("encryptMasterKey() should set ciphertext")
	}

	decrypted, err := decryptMasterKey(encrypted, passphrase)
	if err != nil {
		t.Fatalf("decryptMasterKey() error = %v", err)
	}

	if string(decrypted) != string(testKey) {
		t.Errorf("decryptMasterKey() = %v, want %v", decrypted, testKey)
	}
}

func TestDecryptMasterKey_WrongPassphrase(t *testing.T) {
	testKey := []byte("test-master-key-32-bytes-long!!")

	encrypted, _ := encryptMasterKey(testKey, "correct-passphrase")

	_, err := decryptMasterKey(encrypted, "wrong-passphrase")
	if err == nil {
		t.Error("decryptMasterKey() should fail with wrong passphrase")
	}
}

func TestFileBackend_Name(t *testing.T) {
	b := NewFileBackend()
	if b.Name() != "file" {
		t.Errorf("Name() = %v, want file", b.Name())
	}
}

func TestFileBackend_Available(t *testing.T) {
	b := NewFileBackend()
	if !b.Available() {
		t.Error("FileBackend should always be available")
	}
}
