# Protobuf MCP Server Makefile

.PHONY: help build test lint clean install deps version

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the binary
	@echo "Building protobuf-mcp..."
	@go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o protobuf-mcp ./cmd/protobuf-mcp

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/protobuf-mcp-linux-amd64 ./cmd/protobuf-mcp
	@GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/protobuf-mcp-linux-arm64 ./cmd/protobuf-mcp
	@GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/protobuf-mcp-windows-amd64.exe ./cmd/protobuf-mcp
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/protobuf-mcp-darwin-amd64 ./cmd/protobuf-mcp
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/protobuf-mcp-darwin-arm64 ./cmd/protobuf-mcp

# Test targets
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./internal/mcp/...

# Lint targets
lint: ## Run linters
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

lint-fix: ## Run linters with auto-fix
	@echo "Running linters with auto-fix..."
	@go fmt ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Install targets
install: ## Install the binary
	@echo "Installing protobuf-mcp..."
	@go install -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty) -X main.commit=$(shell git rev-parse HEAD) -X main.date=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)" ./cmd/protobuf-mcp

# Dependency targets
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f protobuf-mcp
	@rm -rf dist/
	@rm -f coverage.out coverage.html

version: ## Show version information
	@echo "Version: $(shell git describe --tags --always --dirty)"
	@echo "Commit: $(shell git rev-parse HEAD)"
	@echo "Date: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# Development targets
dev: build ## Build and run in development mode
	@echo "Starting development server..."
	@./protobuf-mcp server

# CI targets
ci: deps lint test ## Run CI pipeline locally
	@echo "CI pipeline completed successfully"

# Release targets
release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@git diff --exit-code
	@go mod tidy
	@git diff --exit-code go.mod go.sum
	@echo "Ready for release!"

# Default target
.DEFAULT_GOAL := help
