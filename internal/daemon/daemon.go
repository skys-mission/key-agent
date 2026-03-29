// Package daemon provides the daemon lifecycle management.
package daemon

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/skys-mission/key-agent/internal/common"
	"github.com/skys-mission/key-agent/internal/config"
	"github.com/skys-mission/key-agent/internal/logger"
	"github.com/skys-mission/key-agent/internal/mcp"
	"github.com/skys-mission/key-agent/internal/server"
	"github.com/skys-mission/key-agent/internal/storage"
	"github.com/skys-mission/key-agent/internal/storage/bolt"
	"github.com/skys-mission/key-agent/internal/storage/crypto"
	"github.com/skys-mission/key-agent/internal/storage/masterkey"
)

// Run starts the key-agent daemon.
func Run(ctx context.Context, cfg *config.Config) error {
	// Initialize logger
	if err := logger.Init(&cfg.Logging); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("Starting key-agent daemon", "addr", cfg.Server.Addr)

	// Initialize master key provider
	provider := masterkey.DefaultProvider()
	masterKey, err := provider.GetOrCreate()
	if err != nil {
		return err
	}

	// Create encryptor
	encryptor, err := crypto.NewEncryptor(masterKey)
	if err != nil {
		return err
	}
	defer encryptor.Destroy()

	// Initialize storage
	store, err := bolt.New(cfg.DataPath(), encryptor)
	if err != nil {
		return err
	}
	defer store.Close()

	// Initialize tokens if needed
	if err := initTokens(store, cfg.Storage.DataDir); err != nil {
		return fmt.Errorf("failed to initialize tokens: %w", err)
	}

	// Create MCP server if enabled
	var mcpServer *mcp.Server
	if cfg.MCP.Enabled {
		mcpServer = mcp.NewServer(store, "1.0.0")
		logger.Info("MCP server enabled", "endpoint", cfg.MCP.Endpoint)
	}

	// Create HTTP router
	handler := server.Router(store, mcpServer, "1.0.0")

	// Create and start HTTP server
	srv := server.New(cfg.Server.Addr, handler)

	// Handle shutdown
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down key-agent daemon")
		return srv.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// initTokens ensures at least one token exists for authentication.
func initTokens(store *bolt.Store, dataDir string) error {
	tokens, err := store.ListTokens(context.Background())
	if err != nil {
		return err
	}

	if len(tokens) > 0 {
		return nil
	}

	// Generate root token
	token, err := generateRootToken()
	if err != nil {
		return err
	}

	tokenHash := common.HashToken(token)
	now := time.Now()
	meta := &storage.TokenMeta{
		Name:      "root",
		Type:      "client",
		CreatedAt: now,
	}

	if err := store.SetToken(context.Background(), tokenHash, meta); err != nil {
		return err
	}

	// Save token to file for CLI
	tokenPath := filepath.Join(dataDir, "token")
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		logger.Warn("Failed to save token file", "path", tokenPath, "error", err)
	}

	fmt.Println("========================================")
	fmt.Println("Root token generated (save this token):")
	fmt.Printf("%s\n", token)
	fmt.Println("========================================")
	fmt.Println("Use this token for API authentication.")
	fmt.Println()

	return nil
}

// generateRootToken generates a secure random token.
func generateRootToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ka_" + hex.EncodeToString(bytes), nil
}
