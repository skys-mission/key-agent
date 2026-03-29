.PHONY: build build-all test clean lint fmt help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
help:
	@echo "key-agent - Local Key-Value Agent Daemon"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build binaries"
	@echo "  make build-all    Build for all platforms"
	@echo "  make test         Run tests"
	@echo "  make lint         Run linter"
	@echo "  make fmt          Format code"
	@echo "  make clean        Clean build artifacts"

# Build binaries
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent ./cmd/key-agent
	CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl ./cmd/keyctl

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent-darwin-amd64 ./cmd/key-agent
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent-darwin-arm64 ./cmd/key-agent
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent-linux-amd64 ./cmd/key-agent
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent-linux-arm64 ./cmd/key-agent
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/key-agent-windows-amd64.exe ./cmd/key-agent
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl-darwin-amd64 ./cmd/keyctl
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl-darwin-arm64 ./cmd/keyctl
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl-linux-amd64 ./cmd/keyctl
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl-linux-arm64 ./cmd/keyctl
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/keyctl-windows-amd64.exe ./cmd/keyctl

# Run tests
test:
	go test -race -cover -count=3 -shuffle=on -timeout=5m ./...

# Run integration tests
test-integration:
	go test -tags=integration -race -cover -timeout=10m ./tests/integration/...

# Run E2E tests
test-e2e:
	go test -tags=e2e -race -timeout=10m ./tests/e2e/...

# Run linter
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf ~/.key-agent/data/*.db 2>/dev/null || true

# Run daemon locally (for development)
run:
	go run ./cmd/key-agent

# Install locally
install:
	CGO_ENABLED=0 go install $(LDFLAGS) ./cmd/key-agent
	CGO_ENABLED=0 go install $(LDFLAGS) ./cmd/keyctl
