.PHONY: all server client clean docker help

# Default target
all: server client

# Build server
server:
	@echo "🔨 Building CipherWall server..."
	@go build -o cipherwall-server main.go
	@echo "✅ Server built: ./cipherwall-server"

# Build client
client:
	@echo "🔨 Building CipherWall client..."
	@go build -tags client -o cipherwall-client client.go
	@echo "✅ Client built: ./cipherwall-client"

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	@docker build -t cipherwall-vpn:latest -f Dockerfile .
	@echo "✅ Docker image built: cipherwall-vpn:latest"

# Build Dokploy image
docker-dokploy:
	@echo "🐳 Building Dokploy Docker image..."
	@docker build -t cipherwall-vpn:dokploy -f Dockerfile.dokploy .
	@echo "✅ Docker image built: cipherwall-vpn:dokploy"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f cipherwall-server cipherwall-client
	@echo "✅ Clean complete"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	@go mod download
	@echo "✅ Dependencies installed"

# Run server (requires sudo)
run-server: server
	@echo "🚀 Starting server (requires sudo)..."
	@sudo ./cipherwall-server

# Run client (requires sudo and server IP)
run-client: client
	@if [ -z "$(SERVER)" ]; then \
		echo "❌ Please specify SERVER=<ip:port>"; \
		echo "Example: make run-client SERVER=192.168.1.100:1194"; \
		exit 1; \
	fi
	@echo "🚀 Connecting to $(SERVER) (requires sudo)..."
	@sudo ./cipherwall-client -server $(SERVER)

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@echo "✅ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "🔍 Linting code..."
	@golangci-lint run ./...

# Generate secure PSK
generate-psk:
	@echo "🔑 Generating secure 32-byte PSK..."
	@openssl rand -base64 32 | cut -c1-32

# Show help
help:
	@echo "CipherWall VPN - Makefile Commands"
	@echo "===================================="
	@echo ""
	@echo "Building:"
	@echo "  make all           - Build both server and client"
	@echo "  make server        - Build server only"
	@echo "  make client        - Build client only"
	@echo "  make docker        - Build Docker image"
	@echo "  make docker-dokploy - Build Dokploy Docker image"
	@echo ""
	@echo "Running:"
	@echo "  make run-server    - Run server (requires sudo)"
	@echo "  make run-client SERVER=ip:port - Run client (requires sudo)"
	@echo ""
	@echo "Development:"
	@echo "  make deps          - Install dependencies"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo "  make lint          - Lint code"
	@echo "  make clean         - Remove build artifacts"
	@echo ""
	@echo "Utilities:"
	@echo "  make generate-psk  - Generate a secure PSK"
	@echo "  make help          - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make server"
	@echo "  sudo ./cipherwall-server"
	@echo ""
	@echo "  make client"
	@echo "  sudo ./cipherwall-client -server 192.168.1.100:1194"
	@echo ""
	@echo "  make run-client SERVER=192.168.1.100:1194"
