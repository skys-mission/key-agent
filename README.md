<p align="center">
  <img src="https://img.shields.io/github/v/release/skys-mission/key-agent?include_prereleases" alt="Release">
  <img src="https://img.shields.io/github/go-mod/go-version/skys-mission/key-agent" alt="Go Version">
  <img src="https://img.shields.io/github/license/skys-mission/key-agent" alt="License">
  <img src="https://img.shields.io/github/actions/workflow/status/skys-mission/key-agent/ci.yml?branch=main" alt="CI Status">
  <img src="https://codecov.io/gh/skys-mission/key-agent/branch/main/graph/badge.svg" alt="Coverage">
</p>

<h1 align="center">🔑 Key Agent</h1>

<p align="center">
  <strong>A lightweight, secure local key-value and secrets management daemon</strong><br>
  <sub>For developers, DevOps, and AI agents</sub>
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-sdk">SDK</a> •
  <a href="docs/README_CN.md">中文文档</a>
</p>

---

## ✨ Features

- 🔐 **Encrypted Storage** - AES-256-GCM encryption with master key stored in OS keyring, TPM, or file
- 🤖 **MCP Support** - Native integration with AI agents via [Model Context Protocol](https://modelcontextprotocol.io/)
- 🖥️ **CLI Tool** - Simple and intuitive command-line interface (`keyctl`)
- 🔌 **Go SDK** - Programmatic access from Go applications
- 📝 **Structured Logging** - JSON logging with rotation and size limits
- 🚀 **Single Binary** - Zero external dependencies, easy deployment
- 🔄 **Token Management** - Create and manage access tokens with expiration

## 📦 Installation

### Docker (Recommended for Quick Start)

```bash
# Using Docker Compose
git clone https://github.com/skys-mission/key-agent.git
cd key-agent

# Generate a master key (recommended)
export KEY_AGENT_MASTER_KEY=$(openssl rand -base64 32)

# Start the service
docker compose up -d

# View logs to get root token
docker logs key-agent
```

**Environment Variables:**

| Variable | Description |
|----------|-------------|
| `KEY_AGENT_MASTER_KEY` | Base64-encoded 32-byte master key (recommended for Docker) |
| `KEY_AGENT_PASSPHRASE` | Passphrase for encrypted master key file |

### From Release

```bash
# macOS / Linux
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/key-agent-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o key-agent
chmod +x key-agent
sudo mv key-agent /usr/local/bin/

# CLI tool
curl -sL https://github.com/skys-mission/key-agent/releases/latest/download/keyctl-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o keyctl
chmod +x keyctl
sudo mv keyctl /usr/local/bin/
```

### From Source

```bash
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
make build
sudo make install
```

## 🚀 Quick Start

### Option 1: Docker (Fastest)

```bash
# Clone and start
git clone https://github.com/skys-mission/key-agent.git
cd key-agent
docker compose up -d

# Check service health
curl http://127.0.0.1:8080/health

# View logs to get root token
docker logs key-agent
```

### Option 2: Binary

#### 1. Start the Daemon

```bash
key-agent
```

On first run, a root token is generated and displayed. **Save this token securely!**

```
========================================
Root token generated (save this token):
ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
========================================
```

#### 2. Save Your Token

```bash
# Save token for CLI usage
keyctl token save ka_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

#### 3. Store and Retrieve Values

```bash
# Store a key-value pair
keyctl kv set app/database/host "localhost"
keyctl kv set app/database/port "5432"

# Retrieve a value
keyctl kv get app/database/host
# Output: localhost

# List all keys
keyctl kv list
```

#### 4. Store Secrets

```bash
# Store an API key
keyctl secret set openai/api_key "sk-xxxxxxxx" --type api_key

# Store a password
keyctl secret set db/postgres/password "mysecretpass" --type password

# Retrieve a secret
keyctl secret get openai/api_key
```

## 📖 Usage

### Key-Value Operations

```bash
# Set a value with metadata
keyctl kv set config/timeout "30s"

# Get raw value only
keyctl kv get config/timeout --raw

# List keys with prefix
keyctl kv list --prefix app/

# Delete a key
keyctl kv delete config/timeout
```

### Secret Operations

```bash
# Available secret types: password, api_key, certificate, private_key, token, other
keyctl secret set aws/access_key "AKIAxxxx" --type api_key
keyctl secret set ssh/private_key "-----BEGIN..." --type private_key

# List secrets
keyctl secret list
```

### Token Management

```bash
# Create a new token
keyctl token create --name "my-app" --type client --expires-in 24h

# Save token to file
keyctl token save ka_xxxx
```

## 🔌 SDK

Key Agent provides a Go SDK for programmatic access:

```go
package main

import (
    "fmt"
    "github.com/skys-mission/keysdk"
)

func main() {
    client := keysdk.NewClient(&keysdk.Config{
        BaseURL: "http://127.0.0.1:8080",
        Token:   "your-token-here",
    })

    // Set a value
    entry, _ := client.SetKV("my-key", &keysdk.SetKVOptions{
        Value: "my-value",
    })
    fmt.Printf("Created: %s\n", entry.Key)

    // Get a value
    entry, _ = client.GetKV("my-key")
    fmt.Printf("Value: %s\n", entry.Value)

    // Store a secret
    secret, _ := client.SetSecret("db-password", &keysdk.SetSecretOptions{
        Value: "super-secret",
        Type:  keysdk.SecretTypePassword,
    })
    fmt.Printf("Secret type: %s\n", secret.Type)
}
```

## 🤖 MCP Integration

Key Agent supports the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) for AI agent integration. Enable it in your config:

```yaml
mcp:
  enabled: true
  endpoint: /mcp
```

Configure your AI assistant to connect to `http://localhost:8080/mcp` with a valid token.

## ⚙️ Configuration

Configuration file: `~/.key-agent/config.yaml`

```yaml
server:
  addr: 127.0.0.1:8080

storage:
  data_dir: ~/.key-agent/data
  db_name: key-agent.db

security:
  master_key_backend: auto  # auto, keyring, tpm, file

logging:
  level: info
  format: json
  file: ""
  max_size: 100       # MB
  max_backups: 3
  max_age: 30         # days
  compress: true

mcp:
  enabled: true
  endpoint: /mcp
```

## 🛡️ Security

- **Encryption**: All data is encrypted at rest using AES-256-GCM
- **Master Key**: Stored securely in OS keyring (Keychain on macOS, Secret Service on Linux)
- **Token-based Auth**: All API operations require valid tokens
- **No Network Exposure**: By default binds to localhost only

## 📚 Documentation

| Document | Description |
|----------|-------------|
| [User Guide (EN)](docs/USER_GUIDE.md) | Detailed usage instructions |
| [User Guide (CN)](docs/USER_GUIDE_CN.md) | 详细使用指南 |
| [API Reference](docs/API.md) | HTTP API documentation |
| [Development Guide](docs/DEVELOPMENT.md) | Contributing and development |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

```bash
# Run tests
make test

# Run linter
make lint

# Build
make build
```

## 📄 License

[MIT License](LICENSE) 

---

<p align="center">
  95% Vibe Coding ✨ 
</p>
