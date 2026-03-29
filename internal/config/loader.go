// Package config provides configuration loading with priority chain.
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load loads configuration with priority: CLI > File > Env > Default.
func Load() (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Parse CLI flags first (to get config file path and version flag)
	var (
		configFile string
		showHelp   bool
	)

	// Define flags
	flag.StringVar(&configFile, "config", "", "Path to config file")
	flag.StringVar(&cfg.Server.Addr, "addr", cfg.Server.Addr, "Server address (host:port)")
	flag.BoolVar(&cfg.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	// Parse flags
	flag.Parse()

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Load from config file (if exists)
	if configFile == "" {
		configFile = Path()
	}
	if err := loadFromFile(cfg, configFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Load from environment variables (override file)
	loadFromEnv(cfg)

	// Re-parse flags (override everything)
	flag.Parse()

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file.
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(cfg *Config) {
	// Server
	if v := os.Getenv("KEY_AGENT_ADDR"); v != "" {
		cfg.Server.Addr = v
	}

	// Storage
	if v := os.Getenv("KEY_AGENT_DATA_DIR"); v != "" {
		cfg.Storage.DataDir = v
	}
	if v := os.Getenv("KEY_AGENT_DB_NAME"); v != "" {
		cfg.Storage.DBName = v
	}

	// Security
	if v := os.Getenv("KEY_AGENT_MASTER_KEY_BACKEND"); v != "" {
		cfg.Security.MasterKeyBackend = v
	}

	// Logging
	if v := os.Getenv("KEY_AGENT_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("KEY_AGENT_LOG_FILE"); v != "" {
		cfg.Logging.File = v
	}

	// MCP
	if v := os.Getenv("KEY_AGENT_MCP_ENABLED"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.MCP.Enabled = b
		}
	}
	if v := os.Getenv("KEY_AGENT_MCP_ENDPOINT"); v != "" {
		cfg.MCP.Endpoint = v
	}
}

// Save saves the configuration to a YAML file.
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate server address
	if c.Server.Addr == "" {
		return fmt.Errorf("server address is required")
	}
	if !strings.Contains(c.Server.Addr, ":") {
		return fmt.Errorf("server address must be in host:port format")
	}

	// Validate master key backend
	validBackends := map[string]bool{
		"auto":    true,
		"keyring": true,
		"tpm":     true,
		"file":    true,
	}
	if !validBackends[c.Security.MasterKeyBackend] {
		return fmt.Errorf("invalid master key backend: %s", c.Security.MasterKeyBackend)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s", c.Logging.Level)
	}

	return nil
}
