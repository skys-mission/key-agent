package common

import (
	"testing"
)

func TestHashToken(t *testing.T) {
	token := "ka_test123"
	hash := HashToken(token)

	if hash == "" {
		t.Error("HashToken() returned empty string")
	}

	if hash == token {
		t.Error("HashToken() should not return the original token")
	}

	// Same token should produce same hash
	hash2 := HashToken(token)
	if hash != hash2 {
		t.Error("HashToken() should produce consistent hashes")
	}

	// Different tokens should produce different hashes
	hash3 := HashToken("ka_different")
	if hash == hash3 {
		t.Error("HashToken() should produce different hashes for different tokens")
	}
}

func TestHashToken_Empty(t *testing.T) {
	hash := HashToken("")
	if hash == "" {
		t.Error("HashToken(\"\") should still return a hash")
	}
	// SHA-256 of empty string is a known value
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expectedHash {
		t.Errorf("HashToken(\"\") = %v, want %v", hash, expectedHash)
	}
}
