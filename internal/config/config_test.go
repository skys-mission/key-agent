package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Addr != "127.0.0.1:8080" {
		t.Errorf("DefaultConfig() Server.Addr = %v, want 127.0.0.1:8080", cfg.Server.Addr)
	}
	if cfg.Storage.DBName != "key-agent.db" {
		t.Errorf("DefaultConfig() Storage.DBName = %v, want key-agent.db", cfg.Storage.DBName)
	}
	if cfg.Security.MasterKeyBackend != "auto" {
		t.Errorf("DefaultConfig() Security.MasterKeyBackend = %v, want auto", cfg.Security.MasterKeyBackend)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("DefaultConfig() Logging.Level = %v, want info", cfg.Logging.Level)
	}
	if !cfg.MCP.Enabled {
		t.Error("DefaultConfig() MCP.Enabled should be true")
	}
	if cfg.MCP.Endpoint != "/mcp" {
		t.Errorf("DefaultConfig() MCP.Endpoint = %v, want /mcp", cfg.MCP.Endpoint)
	}
}

func TestDataPath(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			DataDir: "/data",
			DBName:  "test.db",
		},
	}

	expected := "/data/test.db"
	if got := cfg.DataPath(); got != expected {
		t.Errorf("DataPath() = %v, want %v", got, expected)
	}
}

func TestTokenPath(t *testing.T) {
	cfg := &Config{
		Storage: StorageConfig{
			DataDir: "/data",
		},
	}

	expected := "/data/token"
	if got := cfg.TokenPath(); got != expected {
		t.Errorf("TokenPath() = %v, want %v", got, expected)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty server addr",
			cfg: &Config{
				Server:   ServerConfig{Addr: ""},
				Security: SecurityConfig{MasterKeyBackend: "auto"},
				Logging:  LoggingConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "invalid server addr format",
			cfg: &Config{
				Server:   ServerConfig{Addr: "localhost"},
				Security: SecurityConfig{MasterKeyBackend: "auto"},
				Logging:  LoggingConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "invalid master key backend",
			cfg: &Config{
				Server:   ServerConfig{Addr: "127.0.0.1:8080"},
				Security: SecurityConfig{MasterKeyBackend: "invalid"},
				Logging:  LoggingConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Server:   ServerConfig{Addr: "127.0.0.1:8080"},
				Security: SecurityConfig{MasterKeyBackend: "auto"},
				Logging:  LoggingConfig{Level: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "key-agent-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Server: ServerConfig{
			Addr: "127.0.0.1:9090",
		},
		Storage: StorageConfig{
			DataDir: "/custom/data",
			DBName:  "custom.db",
		},
		Security: SecurityConfig{
			MasterKeyBackend: "file",
		},
		Logging: LoggingConfig{
			Level: "debug",
			File:  "/var/log/key-agent.log",
		},
		MCP: MCPConfig{
			Enabled:  false,
			Endpoint: "/custom-mcp",
		},
	}

	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test Save
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test loadFromFile
	loaded := DefaultConfig()
	if err := loadFromFile(loaded, configPath); err != nil {
		t.Fatalf("loadFromFile() error: %v", err)
	}

	if loaded.Server.Addr != cfg.Server.Addr {
		t.Errorf("loaded Server.Addr = %v, want %v", loaded.Server.Addr, cfg.Server.Addr)
	}
	if loaded.Storage.DataDir != cfg.Storage.DataDir {
		t.Errorf("loaded Storage.DataDir = %v, want %v", loaded.Storage.DataDir, cfg.Storage.DataDir)
	}
	if loaded.Security.MasterKeyBackend != cfg.Security.MasterKeyBackend {
		t.Errorf("loaded Security.MasterKeyBackend = %v, want %v", loaded.Security.MasterKeyBackend, cfg.Security.MasterKeyBackend)
	}
	if loaded.Logging.Level != cfg.Logging.Level {
		t.Errorf("loaded Logging.Level = %v, want %v", loaded.Logging.Level, cfg.Logging.Level)
	}
	if loaded.MCP.Enabled != cfg.MCP.Enabled {
		t.Errorf("loaded MCP.Enabled = %v, want %v", loaded.MCP.Enabled, cfg.MCP.Enabled)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"KEY_AGENT_ADDR":               "127.0.0.1:9999",
		"KEY_AGENT_DATA_DIR":           "/env/data",
		"KEY_AGENT_DB_NAME":            "env.db",
		"KEY_AGENT_MASTER_KEY_BACKEND": "tpm",
		"KEY_AGENT_LOG_LEVEL":          "debug",
		"KEY_AGENT_LOG_FILE":           "/env/log.txt",
		"KEY_AGENT_MCP_ENABLED":        "false",
		"KEY_AGENT_MCP_ENDPOINT":       "/env-mcp",
	}

	// Set env vars
	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	cfg := DefaultConfig()
	loadFromEnv(cfg)

	if cfg.Server.Addr != "127.0.0.1:9999" {
		t.Errorf("loadFromEnv() Server.Addr = %v", cfg.Server.Addr)
	}
	if cfg.Storage.DataDir != "/env/data" {
		t.Errorf("loadFromEnv() Storage.DataDir = %v", cfg.Storage.DataDir)
	}
	if cfg.Storage.DBName != "env.db" {
		t.Errorf("loadFromEnv() Storage.DBName = %v", cfg.Storage.DBName)
	}
	if cfg.Security.MasterKeyBackend != "tpm" {
		t.Errorf("loadFromEnv() Security.MasterKeyBackend = %v", cfg.Security.MasterKeyBackend)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("loadFromEnv() Logging.Level = %v", cfg.Logging.Level)
	}
	if cfg.Logging.File != "/env/log.txt" {
		t.Errorf("loadFromEnv() Logging.File = %v", cfg.Logging.File)
	}
	if cfg.MCP.Enabled {
		t.Error("loadFromEnv() MCP.Enabled should be false")
	}
	if cfg.MCP.Endpoint != "/env-mcp" {
		t.Errorf("loadFromEnv() MCP.Endpoint = %v", cfg.MCP.Endpoint)
	}
}

func TestLoadFromFileNonExistent(t *testing.T) {
	cfg := DefaultConfig()
	err := loadFromFile(cfg, "/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("loadFromFile() should return error for non-existent file")
	}
}
