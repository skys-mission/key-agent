# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- MCP (Model Context Protocol) support for AI agent integration
- **Go SDK (keysdk)**
  - `GetStringOr(key, default)` - Get KV value with default fallback
  - `GetSecretStringOr(key, default)` - Get secret value with default fallback
- **CLI Configuration**
  - Persistent configuration file at `~/.skys-mission/key-agent/keyctl.yaml`
  - `keyctl config set/get/list` commands for managing settings
  - Configuration priority: CLI flags > Environment variables > Config file
  - Environment variables: `KEY_AGENT_ADDR`, `KEY_AGENT_TOKEN`

### Changed

- **BREAKING**: Configuration directory changed from `~/.key-agent/` to `~/.skys-mission/key-agent/`
  - Daemon config: `~/.skys-mission/key-agent/config.yaml`
  - CLI config: `~/.skys-mission/key-agent/keyctl.yaml`
  - Data directory: `~/.skys-mission/key-agent/data/`
  - Token file: `~/.skys-mission/key-agent/data/token`

## [0.1.0] - 2024-03-30

### Added

- **Core Features**
  - Encrypted key-value storage with AES-256-GCM
  - Secret management with typed entries (password, api_key, certificate, private_key, token, other)
  - Token-based authentication with expiration support
  - Single binary deployment with zero external dependencies

- **CLI Tool (keyctl)**
  - KV operations: `get`, `set`, `delete`, `list`
  - Secret operations: `get`, `set`, `delete`, `list` with type support
  - Token management: `create`, `save`
  - Configurable server address and token file location

- **HTTP API**
  - RESTful endpoints for KV and Secret operations
  - Token creation endpoint
  - Health check endpoint
  - Bearer token authentication

- **Go SDK (keysdk)**
  - Full API coverage
  - Simple client configuration
  - Error handling with typed errors

- **Security**
  - Master key storage in OS keyring (Keychain/Secret Service)
  - TPM support for Linux
  - File-based fallback for restricted environments
  - Token hashing for secure storage

- **Logging**
  - Structured JSON logging
  - Log rotation with size limits
  - Configurable log levels
  - HTTP request logging middleware

- **Configuration**
  - YAML configuration file support
  - Environment-based defaults
  - Configurable server, storage, security, and logging options

### Documentation

- User guide (English and Chinese)
- API reference
- Development guide
- Contributing guidelines

### CI/CD

- GitHub Actions workflow for testing and releases
- Multi-platform binary builds (Darwin, Linux, Windows)
- Automated release artifacts

[Unreleased]: https://github.com/skys-mission/key-agent/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/skys-mission/key-agent/releases/tag/v0.1.0
