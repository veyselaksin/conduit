# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git build-base linux-headers

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cipherwall-server main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    iptables \
    iproute2 \
    ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cipherwall-server .
COPY setup-server.sh .

# Make scripts executable
RUN chmod +x cipherwall-server setup-server.sh

# Expose VPN port
EXPOSE 1194/udp

# Create entrypoint script
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'set -e' >> /app/entrypoint.sh && \
    echo 'echo "🛡️  CipherWall VPN Server - Docker Container"' >> /app/entrypoint.sh && \
    echo 'echo "============================================"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Run setup script' >> /app/entrypoint.sh && \
    echo './setup-server.sh' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo 'echo ""' >> /app/entrypoint.sh && \
    echo 'echo "🚀 Starting CipherWall server..."' >> /app/entrypoint.sh && \
    echo './cipherwall-server' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Run as root (required for TUN interface and iptables)
ENTRYPOINT ["/app/entrypoint.sh"]
