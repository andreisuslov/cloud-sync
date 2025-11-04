# Cloud Sync Makefile
# Build and install the csync command

# Binary name
BINARY_NAME=csync
# Installation directory (in user's PATH)
INSTALL_DIR=$(HOME)/.local/bin
# Source directory
CMD_DIR=./cmd/cloud-sync

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(CMD_DIR)
	@echo "✓ Built $(BINARY_NAME)"

# Build with version info
.PHONY: build-version
build-version:
	@echo "Building $(BINARY_NAME) with version info..."
	@go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) $(CMD_DIR)
	@echo "✓ Built $(BINARY_NAME) v$(VERSION)"

# Install the binary to ~/.local/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@echo "Code signing binary..."
	@codesign --force --deep --sign - $(BINARY_NAME) 2>/dev/null || true
	@mkdir -p $(INSTALL_DIR)
	@cp $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_DIR)/$(BINARY_NAME)
	@xattr -cr $(INSTALL_DIR)/$(BINARY_NAME) 2>/dev/null || true
	@echo "✓ Installed $(BINARY_NAME) to $(INSTALL_DIR)"
	@echo ""
	@echo "Make sure $(INSTALL_DIR) is in your PATH."
	@echo "Add this to your ~/.zshrc or ~/.bash_profile if needed:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""
	@echo ""
	@echo "Then run: source ~/.zshrc (or restart your terminal)"
	@echo "You can now use: csync"

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Uninstalled $(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -f cloud-sync cloud-sync-debug cloud-sync-test
	@rm -f cloud-sync-amd64 cloud-sync-arm64
	@echo "✓ Cleaned"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

# Run the application (for development)
.PHONY: run
run: build
	@./$(BINARY_NAME)

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@go vet ./...
	@echo "✓ Linted"

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# Build for multiple architectures
.PHONY: build-all
build-all:
	@echo "Building for all architectures..."
	@GOARCH=amd64 go build -o $(BINARY_NAME)-amd64 $(CMD_DIR)
	@GOARCH=arm64 go build -o $(BINARY_NAME)-arm64 $(CMD_DIR)
	@echo "✓ Built $(BINARY_NAME)-amd64 and $(BINARY_NAME)-arm64"

# Help
.PHONY: help
help:
	@echo "Cloud Sync - Makefile commands:"
	@echo ""
	@echo "  make build          - Build the csync binary"
	@echo "  make install        - Build and install csync to ~/.local/bin"
	@echo "  make uninstall      - Remove csync from ~/.local/bin"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make run            - Build and run the application"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make build-all      - Build for all architectures"
	@echo "  make help           - Show this help message"
	@echo ""
