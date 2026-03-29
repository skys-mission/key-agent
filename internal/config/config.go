// Package config provides configuration management for key-agent.
package config

import (
	"os"
	"path/filepath"
)

// Config holds all configuration for key-agent daemon.
type Config struct {
	// Server configuration
	Server ServerConfig `yaml:"server" json:"server"`

	// Storage configuration
	Storage StorageConfig `yaml:"storage" json:"storage"`

	// Security configuration
	Security SecurityConfig `yaml:"security" json:"security"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// MCP configuration
	MCP MCPConfig `yaml:"mcp" json:"mcp"`

	// CLI flags
	ShowVersion bool `yaml:"-" json:"-"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	// Addr is the address to listen on (default: 127.0.0.1:8080)
	Addr string `yaml:"addr" json:"addr"`
}

// StorageConfig holds storage configuration.
type StorageConfig struct {
	// DataDir is the directory for data storage
	DataDir string `yaml:"data_dir" json:"data_dir"`

	// DBName is the database file name
	DBName string `yaml:"db_name" json:"db_name"`
}

// SecurityConfig holds security configuration.
type SecurityConfig struct {
	// MasterKeyBackend specifies the master key storage backend
	// Options: auto, keyring, tpm, file
	MasterKeyBackend string `yaml:"master_key_backend" json:"master_key_backend"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	// Level is the log level: debug, info, warn, error
	Level string `yaml:"level" json:"level"`

	// File is the log file path (empty for stderr)
	File string `yaml:"file" json:"file"`

	// MaxSize is the maximum size in megabytes before rotation (default: 100)
	MaxSize int `yaml:"max_size" json:"max_size"`

	// MaxBackups is the maximum number of old log files to retain (default: 3)
	MaxBackups int `yaml:"max_backups" json:"max_backups"`

	// MaxAge is the maximum number of days to retain old log files (default: 30)
	MaxAge int `yaml:"max_age" json:"max_age"`

	// Compress determines if rotated log files should be compressed
	Compress bool `yaml:"compress" json:"compress"`

	// Format is the log format: json or text (default: json)
	Format string `yaml:"format" json:"format"`
}

// MCPConfig holds MCP configuration.
type MCPConfig struct {
	// Enabled controls whether MCP server is enabled
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Endpoint is the MCP endpoint path (default: /mcp)
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Server: ServerConfig{
			Addr: "127.0.0.1:8080",
		},
		Storage: StorageConfig{
			DataDir: filepath.Join(homeDir, ".key-agent", "data"),
			DBName:  "key-agent.db",
		},
		Security: SecurityConfig{
			MasterKeyBackend: "auto",
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
			Format:     "json",
		},
		MCP: MCPConfig{
			Enabled:  true,
			Endpoint: "/mcp",
		},
	}
}

// DataPath returns the full path to the database file.
func (c *Config) DataPath() string {
	return filepath.Join(c.Storage.DataDir, c.Storage.DBName)
}

// TokenPath returns the path to the token file.
func (c *Config) TokenPath() string {
	return filepath.Join(c.Storage.DataDir, "token")
}

// Path returns the default config file path.
func Path() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".key-agent", "config.yaml")
}
