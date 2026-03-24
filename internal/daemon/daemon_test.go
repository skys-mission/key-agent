package daemon

import (
	"testing"
)

func TestGenerateRootToken(t *testing.T) {
	token, err := generateRootToken()
	if err != nil {
		t.Fatalf("generateRootToken() error = %v", err)
	}

	if token == "" {
		t.Error("generateRootToken() returned empty token")
	}

	// Token should start with "ka_"
	if len(token) < 3 || token[:3] != "ka_" {
		t.Errorf("generateRootToken() token = %v, should start with 'ka_'", token)
	}

	// Token should be long enough (ka_ + 64 hex chars)
	if len(token) < 20 {
		t.Errorf("generateRootToken() token length = %d, should be longer", len(token))
	}
}

func TestGenerateRootToken_Uniqueness(t *testing.T) {
	token1, _ := generateRootToken()
	token2, _ := generateRootToken()

	if token1 == token2 {
		t.Error("generateRootToken() should generate unique tokens")
	}
}
