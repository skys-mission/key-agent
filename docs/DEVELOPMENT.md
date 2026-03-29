# Development Guide

This guide covers development setup, architecture, and best practices for Key Agent.

## Development Setup

### Prerequisites

- **Go 1.22+** - Primary language
- **Make** - Build automation (optional)

### Quick Start

```bash
# Clone repository
git clone https://github.com/skys-mission/key-agent.git
cd key-agent

# Download dependencies
go mod download

# Build
make build

# Run tests
make test

# Run daemon
./bin/key-agent

# Run CLI
./bin/keyctl --help
```

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────┐
│                        keyctl CLI                        │
└─────────────────────────┬───────────────────────────────┘
                          │ HTTP API
                          ▼
┌─────────────────────────────────────────────────────────┐
│                     key-agent Daemon                    │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐ │
│  │ HTTP Server │  │ MCP Server  │  │ Auth Middleware │ │
│  └──────┬──────┘  └──────┬──────┘  └────────┬────────┘ │
│         │                │                   │          │
│         └────────────────┼───────────────────┘          │
│                          ▼                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Storage Layer (BoltDB)              │  │
│  │  ┌────────────┐  ┌──────────┐  ┌─────────────┐  │  │
│  │  │    KV      │  │ Secrets  │  │   Tokens    │  │  │
│  │  └────────────┘  └──────────┘  └─────────────┘  │  │
│  └──────────────────────────────────────────────────┘  │
│                          │                              │
│                          ▼                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Encryption (AES-256-GCM)            │  │
│  └──────────────────────────────────────────────────┘  │
│                          │                              │
│                          ▼                              │
│  ┌──────────────────────────────────────────────────┐  │
│  │              Master Key (Keyring/TPM)            │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Package Structure

| Package | Description |
|---------|-------------|
| `cmd/key-agent` | Daemon entrypoint |
| `cmd/keyctl` | CLI entrypoint |
| `internal/config` | Configuration management |
| `internal/daemon` | Daemon lifecycle |
| `internal/logger` | Structured logging |
| `internal/server` | HTTP server and routing |
| `internal/server/handlers` | Request handlers |
| `internal/server/middleware` | HTTP middleware |
| `internal/storage` | Storage interface |
| `internal/storage/bolt` | BoltDB implementation |
| `internal/storage/crypto` | Encryption utilities |
| `internal/storage/masterkey` | Master key management |
| `internal/mcp` | MCP server |
| `internal/common` | Shared utilities |
| `keysdk` | Go SDK (independent module) |

### Data Flow

1. **Request** → HTTP Server → Router → Middleware → Handler
2. **Handler** → Storage Interface → BoltDB → Encryption Layer
3. **Encryption** → AES-256-GCM using Master Key
4. **Master Key** → OS Keyring / TPM / File

## Key Components

### Configuration

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig
    Storage  StorageConfig
    Security SecurityConfig
    Logging  LoggingConfig
    MCP      MCPConfig
}
```

### Storage Interface

```go
// internal/storage/store.go
type Store interface {
    // KV operations
    GetKV(ctx context.Context, key string) (*Entry, error)
    SetKV(ctx context.Context, entry *Entry) error
    DeleteKV(ctx context.Context, key string) error
    ListKV(ctx context.Context, prefix string) ([]string, error)

    // Secret operations
    GetSecret(ctx context.Context, key string) (*SecretEntry, error)
    SetSecret(ctx context.Context, entry *SecretEntry) error
    DeleteSecret(ctx context.Context, key string) error
    ListSecret(ctx context.Context, prefix string) ([]string, error)

    // Token operations
    GetToken(ctx context.Context, tokenHash string) (*TokenMeta, error)
    SetToken(ctx context.Context, tokenHash string, meta *TokenMeta) error
    DeleteToken(ctx context.Context, tokenHash string) error
    ListTokens(ctx context.Context) ([]string, error)

    Close() error
}
```

### Master Key Management

Master keys are stored using multiple backends:

1. **Keyring** (default on macOS/Linux)
   - macOS: Keychain
   - Linux: Secret Service (GNOME Keyring, KDE Wallet)

2. **TPM** (Linux only)
   - Uses Trusted Platform Module

3. **File** (fallback)
   - Encrypted file storage
   - Less secure, use only when necessary

### Encryption

- **Algorithm**: AES-256-GCM
- **Key Derivation**: Master key directly (no KDF needed as master key is random)
- **Nonce**: Random 12 bytes per encryption

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Run specific package
go test -v ./internal/storage/bolt/...
```

### Integration Tests

```bash
# Run integration tests
go test -v ./test/integration/...
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out
```

## Logging

Key Agent uses structured logging via `log/slog`:

```go
import "github.com/skys-mission/key-agent/internal/logger"

// Initialize
logger.Init(&config.LoggingConfig{
    Level:      "info",
    Format:     "json",
    MaxSize:    100,
    MaxBackups: 3,
    MaxAge:     30,
    Compress:   true,
})

// Log messages
logger.Info("Server started", "addr", addr)
logger.Error("Failed to connect", "error", err)
```

## Adding New Features

### Adding a New API Endpoint

1. **Define handler** in `internal/server/handlers/`

```go
func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle request
}
```

2. **Register route** in `internal/server/routes.go`

```go
mux.Handle("/api/v1/my-endpoint", auth.RequireAuth(myHandler))
```

3. **Add tests** in `internal/server/handlers/my_handler_test.go`

### Adding a New CLI Command

1. **Define command** in `internal/client/commands/`

```go
func init() {
    myCmd := &cobra.Command{
        Use:   "mycommand",
        Short: "My command description",
        Run:   myCommandRun,
    }
    myCmd.Flags().String("option", "", "Option description")
    rootCmd.AddCommand(myCmd)
}

func myCommandRun(cmd *cobra.Command, args []string) {
    // Implementation
}
```

### Adding Storage Features

1. **Update interface** in `internal/storage/store.go`
2. **Implement in** `internal/storage/bolt/store.go`
3. **Add tests** in `internal/storage/bolt/store_test.go`

## Release Checklist

- [ ] Update version in code
- [ ] Update CHANGELOG.md
- [ ] Run full test suite
- [ ] Build all platforms
- [ ] Create git tag
- [ ] Push tag to trigger release workflow

## Troubleshooting

### Build Errors

```bash
# Clear Go cache
go clean -cache

# Re-download dependencies
go mod download
```

### Test Failures

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestMyFunction ./...
```

### Keyring Issues

```bash
# On Linux, ensure secret service is available
dbus-send --session --dest=org.freedesktop.secrets --type=method_call /org/freedesktop/secrets org.freedesktop.DBus.Introspectable.Introspect
```
