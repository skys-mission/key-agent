package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("load non-existent config returns empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Empty(t, cfg.Addr)
		assert.Empty(t, cfg.Token)
		assert.Empty(t, cfg.TokenFile)
	})

	t.Run("load existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		// Create config directory and file
		configDir := filepath.Join(tmpDir, ".skys-mission", "key-agent")
		require.NoError(t, os.MkdirAll(configDir, 0755))

		configContent := `addr: http://localhost:9090
token: test-token-123
token_file: /path/to/token
`
		configPath := filepath.Join(configDir, "keyctl.yaml")
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0600))

		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "http://localhost:9090", cfg.Addr)
		assert.Equal(t, "test-token-123", cfg.Token)
		assert.Equal(t, "/path/to/token", cfg.TokenFile)
	})
}

func TestCLIConfig_Save(t *testing.T) {
	t.Run("save config creates directory and file", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		cfg := &CLIConfig{
			Addr:      "http://localhost:8080",
			Token:     "my-token",
			TokenFile: "~/.token",
		}

		err := cfg.Save()
		require.NoError(t, err)

		// Verify file exists
		configPath := filepath.Join(tmpDir, ".skys-mission", "key-agent", "keyctl.yaml")
		_, err = os.Stat(configPath)
		require.NoError(t, err)

		// Load and verify
		loaded, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, cfg.Addr, loaded.Addr)
		assert.Equal(t, cfg.Token, loaded.Token)
		assert.Equal(t, cfg.TokenFile, loaded.TokenFile)
	})
}

func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	path := ConfigPath()
	assert.Equal(t, filepath.Join(tmpDir, ".skys-mission", "key-agent", "keyctl.yaml"), path)
}
