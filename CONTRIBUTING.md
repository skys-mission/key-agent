# Contributing to Key Agent

Thank you for your interest in contributing to Key Agent! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for all contributors.

## How to Contribute

### Reporting Bugs

1. Check existing issues to avoid duplicates
2. Use the bug report template
3. Include:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Go version)

### Suggesting Features

1. Check existing issues for similar suggestions
2. Use the feature request template
3. Describe the use case and expected behavior

### Submitting Changes

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit with conventional commits
7. Push and create a pull request

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make (optional, for using Makefile)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/key-agent.git
cd key-agent

# Add upstream remote
git remote add upstream https://github.com/skys-mission/key-agent.git

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run locally
./bin/key-agent
```

### Project Structure

```
key-agent/
├── cmd/                    # Application entrypoints
│   ├── key-agent/         # Daemon main
│   └── keyctl/            # CLI main
├── internal/               # Private packages
│   ├── common/            # Shared utilities
│   ├── config/            # Configuration
│   ├── daemon/            # Daemon lifecycle
│   ├── logger/            # Structured logging
│   ├── mcp/               # MCP server
│   ├── server/            # HTTP server
│   │   ├── handlers/      # Request handlers
│   │   └── middleware/    # HTTP middleware
│   └── storage/           # Storage layer
│       ├── bolt/          # BoltDB implementation
│       ├── crypto/        # Encryption
│       └── masterkey/     # Master key management
├── keysdk/                 # Go SDK (independent module)
├── test/                   # Test packages
│   └── integration/       # Integration tests
├── docs/                   # Documentation
└── configs/                # Example configs
```

### Coding Standards

#### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `go fmt` before committing
- Run `go vet` to catch issues

#### Comments

- Exported functions/types must have doc comments
- Comment the "why", not the "what"
- Keep comments concise

```go
// HashToken creates a SHA-256 hash of a token.
// Used for secure token storage - tokens are stored by hash, not plaintext.
func HashToken(token string) string {
    // ...
}
```

#### Error Handling

- Return errors, don't panic
- Wrap errors with context using `fmt.Errorf`
- Use sentinel errors for expected conditions

```go
var ErrNotFound = errors.New("entry not found")

func Get(key string) (*Entry, error) {
    if key == "" {
        return nil, fmt.Errorf("key cannot be empty")
    }
    // ...
}
```

#### Testing

- Write tests for new functionality
- Use table-driven tests
- Test error conditions
- Run tests with race detection

```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Run specific package
go test ./internal/storage/...
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `test`: Tests
- `refactor`: Code refactoring
- `chore`: Maintenance

Examples:
```
feat(api): add token expiration support
fix(storage): handle concurrent access correctly
docs(readme): update installation instructions
test(integration): add KV CRUD tests
```

### Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add tests for new functionality
4. Keep PRs focused (one feature/fix per PR)
5. Reference related issues

### Release Process

Maintainers follow this process:

1. Update version in code
2. Update CHANGELOG.md
3. Create tag: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. GitHub Actions builds and releases

## Getting Help

- Open an issue for questions
- Check existing documentation
- Review closed issues/PRs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
