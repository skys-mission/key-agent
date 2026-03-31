// Package integration provides end-to-end integration tests for key-agent.
package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/skys-mission/key-agent/internal/config"
	"github.com/skys-mission/key-agent/internal/daemon"
	"github.com/skys-mission/key-agent/keysdk"
)

var (
	testDataDir string
	testAddr    string
	testToken   string
)

// TestMain sets up the test environment.
func TestMain(m *testing.M) {
	// Set passphrase for file backend (needed in CI)
	os.Setenv("KEY_AGENT_PASSPHRASE", "test-passphrase-for-ci")

	// Create temp directory
	var err error
	testDataDir, err = os.MkdirTemp("", "key-agent-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(testDataDir)

	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find available port: %v\n", err)
		os.Exit(1)
	}
	testAddr = fmt.Sprintf("http://%s", listener.Addr().String())
	listener.Close()

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Addr: strings.TrimPrefix(testAddr, "http://"),
		},
		Storage: config.StorageConfig{
			DataDir: testDataDir,
			DBName:  "test.db",
		},
		Security: config.SecurityConfig{
			MasterKeyBackend: "file",
		},
		Logging: config.LoggingConfig{
			Level:      "warn",
			Format:     "json",
			MaxSize:    10,
			MaxBackups: 1,
			MaxAge:     1,
		},
		MCP: config.MCPConfig{
			Enabled:  false,
			Endpoint: "/mcp",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start daemon in goroutine
	daemonErr := make(chan error, 1)
	go func() {
		daemonErr <- daemon.Run(ctx, cfg)
	}()

	// Wait for server to be ready (poll health endpoint)
	var ready bool
	for i := 0; i < 30; i++ {
		resp, err := http.Get(testAddr + "/health")
		if err == nil {
			resp.Body.Close()
			ready = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		fmt.Fprintf(os.Stderr, "Server failed to start within 3 seconds\n")
		cancel()
		os.Exit(1)
	}

	// Get root token from token file
	tokenPath := filepath.Join(testDataDir, "token")
	tokenData, err := os.ReadFile(tokenPath)
	if err == nil {
		testToken = strings.TrimSpace(string(tokenData))
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not read token file: %v\n", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cancel()
	<-daemonErr

	os.Exit(code)
}

func getTestClient(t *testing.T) *keysdk.Client {
	if testToken == "" {
		t.Skip("No test token available")
	}
	return keysdk.NewClient(&keysdk.Config{
		BaseURL: testAddr,
		Token:   testToken,
	})
}

// === Health Tests ===

func TestHealth(t *testing.T) {
	c := getTestClient(t)

	health, err := c.Health()
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}

	if health.Status != "healthy" && health.Status != "ok" {
		t.Errorf("Health() status = %v, want healthy or ok", health.Status)
	}
}

// === KV Tests ===

func TestKV_CRUD(t *testing.T) {
	c := getTestClient(t)
	key := "test-kv-key"
	value := "test-value"

	// Set
	entry, err := c.SetKV(key, &keysdk.SetKVOptions{
		Value: value,
		Metadata: map[string]interface{}{
			"description": "test entry",
		},
	})
	if err != nil {
		t.Fatalf("SetKV() error = %v", err)
	}
	if entry.Key != key {
		t.Errorf("SetKV() key = %v, want %v", entry.Key, key)
	}
	if entry.Value != value {
		t.Errorf("SetKV() value = %v, want %v", entry.Value, value)
	}

	// Get
	entry, err = c.GetKV(key)
	if err != nil {
		t.Fatalf("GetKV() error = %v", err)
	}
	if entry.Value != value {
		t.Errorf("GetKV() value = %v, want %v", entry.Value, value)
	}

	// List
	keys, err := c.ListKV("")
	if err != nil {
		t.Fatalf("ListKV() error = %v", err)
	}
	if len(keys) == 0 {
		t.Error("ListKV() returned no keys")
	}

	// Delete
	err = c.DeleteKV(key)
	if err != nil {
		t.Fatalf("DeleteKV() error = %v", err)
	}

	// Verify deleted
	_, err = c.GetKV(key)
	if err == nil {
		t.Error("GetKV() should fail after delete")
	}
}

func TestKV_Prefix(t *testing.T) {
	c := getTestClient(t)

	// Create entries with prefix
	prefix := "prefix-test/"
	keys := []string{prefix + "a", prefix + "b", prefix + "c"}

	for _, k := range keys {
		_, err := c.SetKV(k, &keysdk.SetKVOptions{Value: "value"})
		if err != nil {
			t.Fatalf("SetKV(%s) error = %v", k, err)
		}
	}

	// List with prefix
	list, err := c.ListKV(prefix)
	if err != nil {
		t.Fatalf("ListKV() error = %v", err)
	}

	if len(list) < 3 {
		t.Errorf("ListKV() returned %d keys, want at least 3", len(list))
	}

	// Cleanup
	for _, k := range keys {
		c.DeleteKV(k)
	}
}

// === Secret Tests ===

func TestSecret_CRUD(t *testing.T) {
	c := getTestClient(t)
	key := "test-secret-key"

	// Set
	entry, err := c.SetSecret(key, &keysdk.SetSecretOptions{
		Value: "secret-value",
		Type:  keysdk.SecretTypePassword,
		Metadata: map[string]interface{}{
			"username": "testuser",
		},
	})
	if err != nil {
		t.Fatalf("SetSecret() error = %v", err)
	}
	if entry.Key != key {
		t.Errorf("SetSecret() key = %v, want %v", entry.Key, key)
	}
	if entry.Type != keysdk.SecretTypePassword {
		t.Errorf("SetSecret() type = %v, want %v", entry.Type, keysdk.SecretTypePassword)
	}

	// Get
	entry, err = c.GetSecret(key)
	if err != nil {
		t.Fatalf("GetSecret() error = %v", err)
	}
	if entry.Value != "secret-value" {
		t.Errorf("GetSecret() value = %v, want secret-value", entry.Value)
	}

	// List
	keys, err := c.ListSecret("")
	if err != nil {
		t.Fatalf("ListSecret() error = %v", err)
	}
	if len(keys) == 0 {
		t.Error("ListSecret() returned no keys")
	}

	// Delete
	err = c.DeleteSecret(key)
	if err != nil {
		t.Fatalf("DeleteSecret() error = %v", err)
	}

	// Verify deleted
	_, err = c.GetSecret(key)
	if err == nil {
		t.Error("GetSecret() should fail after delete")
	}
}

func TestSecret_AllTypes(t *testing.T) {
	c := getTestClient(t)

	types := []keysdk.SecretType{
		keysdk.SecretTypePassword,
		keysdk.SecretTypeAPIKey,
		keysdk.SecretTypeCertificate,
		keysdk.SecretTypePrivateKey,
		keysdk.SecretTypeToken,
		keysdk.SecretTypeOther,
	}

	for i, typ := range types {
		key := fmt.Sprintf("secret-type-%d", i)
		_, err := c.SetSecret(key, &keysdk.SetSecretOptions{
			Value: "test-value",
			Type:  typ,
		})
		if err != nil {
			t.Fatalf("SetSecret() type=%s error = %v", typ, err)
		}

		entry, err := c.GetSecret(key)
		if err != nil {
			t.Fatalf("GetSecret() type=%s error = %v", typ, err)
		}
		if entry.Type != typ {
			t.Errorf("GetSecret() type = %v, want %v", entry.Type, typ)
		}

		c.DeleteSecret(key)
	}
}

// === Token Tests ===

func TestToken_Create(t *testing.T) {
	c := getTestClient(t)

	token, err := c.CreateToken(&keysdk.CreateTokenOptions{
		Name:      "test-token",
		Type:      "client",
		ExpiresIn: "24h",
	})
	if err != nil {
		t.Fatalf("CreateToken() error = %v", err)
	}

	if token.Token == "" {
		t.Error("CreateToken() returned empty token")
	}
	if token.Name != "test-token" {
		t.Errorf("CreateToken() name = %v, want test-token", token.Name)
	}

	// Verify token works
	newClient := keysdk.NewClient(&keysdk.Config{
		BaseURL: testAddr,
		Token:   token.Token,
	})

	_, err = newClient.Health()
	if err != nil {
		t.Errorf("Token validation failed: %v", err)
	}
}

// === Error Handling Tests ===

func TestError_NotFound(t *testing.T) {
	c := getTestClient(t)

	_, err := c.GetKV("nonexistent-key-12345")
	if err == nil {
		t.Error("GetKV() should return error for nonexistent key")
	}
}

func TestError_Unauthorized(t *testing.T) {
	c := keysdk.NewClient(&keysdk.Config{
		BaseURL: testAddr,
		Token:   "invalid-token",
	})

	_, err := c.GetKV("any-key")
	if err == nil {
		t.Error("GetKV() should return error for invalid token")
	}
}

// === HTTP Server Tests ===

func TestHTTP_ServerStart(t *testing.T) {
	// Test that we can make a simple HTTP request
	resp, err := http.Get(testAddr + "/health")
	if err != nil {
		t.Fatalf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health endpoint status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
