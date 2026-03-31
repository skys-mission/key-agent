# Key Agent User Guide

This guide covers all aspects of using Key Agent for managing key-value data and secrets.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Configuration](#configuration)
- [CLI Reference](#cli-reference)
- [HTTP API](#http-api)
- [Go SDK](#go-sdk)
- [MCP Integration](#mcp-integration)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Key Agent is a local daemon that provides secure storage for key-value data and secrets. It features:

- **Encrypted storage** using AES-256-GCM
- **Token-based authentication** for API access
- **CLI tool** (`keyctl`) for easy management
- **Go SDK** for programmatic access
- **MCP support** for AI agent integration

## Installation

### Docker (Recommended)

The easiest way to run Key Agent is using Docker:

```bash
# Clone the repository
git clone https://github.com/skys-mission/key-agent.git
cd key-agent

# Start with Docker Compose
docker compose up -d

# Check service status
docker compose ps

# View logs (contains root token on first run)
docker compose logs key-agent

# Stop service
docker compose down
```

**Docker Configuration:**

The `docker-compose.yml` file includes:
- Volume persistence for data
- Health checks
- File-based master key backend (required in containers)

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `KEY_AGENT_MASTER_KEY_BACKEND` | Set to `file` for containers |
| `KEY_AGENT_PASSPHRASE` | Passphrase for file backend (recommended) |

**Manual Docker Run:**

```bash
# Build image
docker build -t key-agent:latest .

# Run container
docker run -d \
  --name key-agent \
  -p 127.0.0.1:8080:8080 \
  -v key-agent-data:/data \
  -e KEY_AGENT_MASTER_KEY_BACKEND=file \
  -e KEY_AGENT_PASSPHRASE=your-secure-passphrase \
  key-agent:latest
```

### Binary Download

Download the latest release from [GitHub Releases](https://github.com/skys-mission/key-agent/releases):

```bash
# Download daemon
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/key-agent-linux-amd64 -o key-agent
chmod +x key-agent
sudo mv key-agent /usr/local/bin/

# Download CLI
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/keyctl-linux-amd64 -o keyctl
chmod +x keyctl
sudo mv keyctl /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
make build
```

## Configuration

### Default Locations

| File | Location |
|------|----------|
| Config file | `~/.key-agent/config.yaml` |
| Data directory | `~/.key-agent/data/` |
| Token file | `~/.key-agent/data/token` |

### Configuration File

Create `~/.key-agent/config.yaml`:

```yaml
# Server settings
server:
  addr: "127.0.0.1:8080"

# Storage settings
storage:
  data_dir: "~/.key-agent/data"
  db_name: "key-agent.db"

# Security settings
security:
  master_key_backend: "auto"  # auto, keyring, tpm, file

# Logging settings
logging:
  level: "info"           # debug, info, warn, error
  format: "json"          # json or text
  file: ""                # empty for stderr, or path to log file
  max_size: 100           # max size in MB before rotation
  max_backups: 3          # max number of old log files
  max_age: 30             # max days to retain old logs
  compress: true          # compress rotated logs

# MCP settings
mcp:
  enabled: true
  endpoint: "/mcp"
```

### Master Key Backend Options

| Backend | Description | Platform |
|---------|-------------|----------|
| `auto` | Automatically select best available | All |
| `keyring` | OS keyring (Keychain/Secret Service) | macOS, Linux |
| `tpm` | Trusted Platform Module | Linux |
| `file` | Encrypted file (less secure) | All |

## CLI Reference

### Starting the Daemon

```bash
# Start with default config
key-agent

# Start with custom config
key-agent --config /path/to/config.yaml

# Start with custom data directory
key-agent --data-dir /custom/data/path
```

### KV Operations

```bash
# Set a value
keyctl kv set <key> <value>

# Examples
keyctl kv set app/database/host "localhost"
keyctl kv set app/database/port "5432"
keyctl kv set app/feature/flags '{"enabled": true}'

# Get a value
keyctl kv get <key>

# Get raw value (no JSON formatting)
keyctl kv get <key> --raw

# List all keys
keyctl kv list

# List with prefix filter
keyctl kv list --prefix app/database/

# Delete a key
keyctl kv delete <key>
```

### Secret Operations

```bash
# Set a secret
keyctl secret set <key> <value> --type <type>

# Secret types: password, api_key, certificate, private_key, token, other

# Examples
keyctl secret set db/password "mysecretpass" --type password
keyctl secret set openai/api_key "sk-xxx" --type api_key
keyctl secret set aws/access_key "AKIAxxx" --type api_key

# Get a secret
keyctl secret get <key>

# List secrets
keyctl secret list

# List with prefix
keyctl secret list --prefix db/

# Delete a secret
keyctl secret delete <key>
```

### Token Operations

```bash
# Save token from initial setup
keyctl token save <token>

# Create a new token
keyctl token create --name <name> [--type <type>] [--expires-in <duration>]

# Examples
keyctl token create --name "my-app" --type client --expires-in 24h
keyctl token create --name "mcp-agent" --type mcp --expires-in 30d
```

### CLI Options

```bash
# Specify server address
keyctl --addr http://localhost:8080 kv get mykey

# Specify token directly
keyctl --token "ka_xxx" kv get mykey

# Specify token file
keyctl --token-file /path/to/token kv get mykey
```

## HTTP API

### Base URL

```
http://127.0.0.1:8080
```

All API requests require Bearer token authentication:

```
Authorization: Bearer <token>
```

### Endpoints

#### Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

#### KV Operations

```
GET    /api/v1/kv/<key>     # Get value
PUT    /api/v1/kv/<key>     # Set value
DELETE /api/v1/kv/<key>     # Delete key
GET    /api/v1/kv           # List keys (with ?prefix=...)
```

Set KV Request:
```json
{
  "value": "my-value",
  "metadata": {
    "description": "Optional description"
  }
}
```

#### Secret Operations

```
GET    /api/v1/secrets/<key>     # Get secret
PUT    /api/v1/secrets/<key>     # Set secret
DELETE /api/v1/secrets/<key>     # Delete secret
GET    /api/v1/secrets           # List secrets (with ?prefix=...)
```

Set Secret Request:
```json
{
  "value": "secret-value",
  "type": "password",
  "metadata": {
    "description": "Database password"
  }
}
```

#### Token Operations

```
POST /api/v1/token     # Create new token
```

Create Token Request:
```json
{
  "name": "my-app",
  "type": "client",
  "expires_in": "24h"
}
```

## Go SDK

### Installation

```bash
go get github.com/skys-mission/keysdk
```

### Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/skys-mission/keysdk"
)

func main() {
    // Create client
    client := keysdk.NewClient(&keysdk.Config{
        BaseURL: "http://127.0.0.1:8080",
        Token:   "your-token-here",
    })

    // KV operations
    kvExample(client)

    // Secret operations
    secretExample(client)

    // Token operations
    tokenExample(client)
}

func kvExample(client *keysdk.Client) {
    // Set
    entry, err := client.SetKV("my-key", &keysdk.SetKVOptions{
        Value: "my-value",
        Metadata: map[string]interface{}{
            "description": "Example key",
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Set: %+v\n", entry)

    // Get
    entry, err = client.GetKV("my-key")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Get: %+v\n", entry)

    // List
    keys, err := client.ListKV("my-")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Keys: %v\n", keys)

    // Delete
    err = client.DeleteKV("my-key")
    if err != nil {
        panic(err)
    }
}

func secretExample(client *keysdk.Client) {
    // Set secret
    secret, err := client.SetSecret("db-password", &keysdk.SetSecretOptions{
        Value: "super-secret",
        Type:  keysdk.SecretTypePassword,
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("Secret: %+v\n", secret)

    // Get secret
    secret, err = client.GetSecret("db-password")
    if err != nil {
        panic(err)
    }
    fmt.Printf("Secret value: %s\n", secret.Value)
}

func tokenExample(client *keysdk.Client) {
    // Create token
    token, err := client.CreateToken(&keysdk.CreateTokenOptions{
        Name:      "my-app",
        Type:      "client",
        ExpiresIn: "24h",
    })
    if err != nil {
        panic(err)
    }
    fmt.Printf("New token: %s\n", token.Token)
}
```

## MCP Integration

Key Agent supports the Model Context Protocol (MCP) for AI agent integration.

### Enable MCP

In your config:

```yaml
mcp:
  enabled: true
  endpoint: /mcp
```

### Configure AI Assistant

For Claude Desktop or other MCP clients, add to your configuration:

```json
{
  "mcpServers": {
    "key-agent": {
      "url": "http://127.0.0.1:8080/mcp",
      "headers": {
        "Authorization": "Bearer your-token-here"
      }
    }
  }
}
```

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `kv_get` | Get a KV value |
| `kv_set` | Set a KV value |
| `kv_delete` | Delete a KV entry |
| `kv_list` | List KV keys |
| `secret_get` | Get a secret |
| `secret_set` | Set a secret |
| `secret_delete` | Delete a secret |
| `secret_list` | List secret keys |

## Security Best Practices

### Token Management

1. **Save root token securely** - It's only displayed once
2. **Create limited tokens** - Use expiration times for temporary access
3. **Rotate tokens regularly** - Delete old tokens and create new ones

### Network Security

1. **Default binding** - Binds to localhost only (127.0.0.1)
2. **No TLS** - Designed for local use; use reverse proxy for remote access
3. **Firewall rules** - If binding to external interface, restrict access

### Storage Security

1. **Master key** - Use `keyring` backend when available
2. **Backup** - Backup your data directory securely
3. **File permissions** - Ensure `~/.key-agent/` has restrictive permissions (700)

## Troubleshooting

### Daemon won't start

```bash
# Check if port is in use
lsof -i :8080

# Check logs
key-agent --log-level debug
```

### Token not working

```bash
# Verify token is saved
cat ~/.key-agent/data/token

# Try with explicit token
keyctl --token "ka_xxx" kv list
```

### Permission denied

```bash
# Fix permissions
chmod 700 ~/.key-agent
chmod 600 ~/.key-agent/data/token
```

### Keyring issues on Linux

```bash
# Install dbus and secret-service
sudo apt install dbus libsecret-1-0

# Or use file backend
key-agent --master-key-backend file
```

## Getting Help

- **GitHub Issues**: [github.com/skys-mission/key-agent/issues](https://github.com/skys-mission/key-agent/issues)
- **Documentation**: [docs/](https://github.com/skys-mission/key-agent/tree/main/docs)
