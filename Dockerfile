# Build stage
FROM golang:1.22-alpine AS builder

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
    ca-certificates \
    bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/cipherwall-server .

# Make binary executable
RUN chmod +x cipherwall-server

# Expose VPN port
EXPOSE 1194/udp

# Create entrypoint script with embedded setup commands
RUN echo '#!/bin/bash' > /app/entrypoint.sh && \
    echo 'set -e' >> /app/entrypoint.sh && \
    echo 'echo "🛡️  CipherWall VPN Server - Docker Container"' >> /app/entrypoint.sh && \
    echo 'echo "============================================"' >> /app/entrypoint.sh && \
    echo 'echo ""' >> /app/entrypoint.sh && \
    echo 'echo "Setting up NAT and IP forwarding..."' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Enable IP forwarding' >> /app/entrypoint.sh && \
    echo 'sysctl -w net.ipv4.ip_forward=1' >> /app/entrypoint.sh && \
    echo 'echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Detect the default network interface' >> /app/entrypoint.sh && \
    echo 'DEFAULT_IFACE=$(ip route | grep default | awk '\''{print $5}'\'' | head -n1)' >> /app/entrypoint.sh && \
    echo 'if [ -z "$DEFAULT_IFACE" ]; then' >> /app/entrypoint.sh && \
    echo '    DEFAULT_IFACE="eth0"' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo 'echo "Using network interface: $DEFAULT_IFACE"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Setup NAT for VPN traffic' >> /app/entrypoint.sh && \
    echo 'iptables -t nat -A POSTROUTING -s 10.8.0.0/24 -o $DEFAULT_IFACE -j MASQUERADE' >> /app/entrypoint.sh && \
    echo 'iptables -A FORWARD -i tun0 -j ACCEPT' >> /app/entrypoint.sh && \
    echo 'iptables -A FORWARD -o tun0 -j ACCEPT' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo 'echo "✅ Network setup complete"' >> /app/entrypoint.sh && \
    echo 'echo ""' >> /app/entrypoint.sh && \
    echo 'echo "🚀 Starting CipherWall server on UDP port 1194..."' >> /app/entrypoint.sh && \
    echo './cipherwall-server' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Run as root (required for TUN interface and iptables)
ENTRYPOINT ["/app/entrypoint.sh"]
