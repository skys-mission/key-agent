# Build stage
FROM golang:1.25-alpine AS builder

# Build arguments (can be overridden via --build-arg)
ARG GOPROXY=direct
ARG GOSUMDB=sum.golang.org

# Set Go proxy environment
ENV GOPROXY=${GOPROXY}
ENV GOSUMDB=${GOSUMDB}

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files (including local module)
COPY go.mod go.sum ./
COPY keysdk/go.mod keysdk/go.sum ./keysdk/

# Download dependencies (try direct, fallback to goproxy.cn)
RUN go mod download || \
    (echo "Direct download failed, trying goproxy.cn..." && \
     GOPROXY=https://goproxy.cn,direct go mod download)

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o /build/key-agent ./cmd/key-agent

# Runtime stage
FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 keyagent && \
    adduser -u 1000 -G keyagent -s /bin/sh -D keyagent

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/key-agent /usr/local/bin/key-agent
RUN chmod +x /usr/local/bin/key-agent

# Create data directory
RUN mkdir -p /data && chown -R keyagent:keyagent /data

# Set environment
ENV HOME=/home/keyagent
ENV KEY_AGENT_DATA_DIR=/data

# Switch to non-root user
USER keyagent

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:8080/health || exit 1

# Run daemon
CMD ["key-agent"]
