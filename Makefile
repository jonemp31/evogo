# Evolution API Go Makefile

.PHONY: help build run test clean docker-build docker-run migrate

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build the application
build: ## Build the application
	@echo "Building Evolution API Go..."
	@go build -o bin/evolution-api-go ./cmd/server
	@echo "Build complete!"

# Run the application
run: ## Run the application
	@echo "Starting Evolution API Go..."
	@go run ./cmd/server

# Run tests
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete!"

# Build Docker image
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t evolution-api-go:latest .
	@echo "Docker build complete!"

# Run with Docker Compose
docker-run: ## Run with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "Services started!"

# Stop Docker Compose services
docker-stop: ## Stop Docker Compose services
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped!"

# Run database migrations
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@psql $(DATABASE_URL) -f migrations/001_create_instances_table.sql
	@echo "Migrations complete!"

# Install dependencies
deps: ## Install dependencies
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed!"

# Format code
fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

# Lint code
lint: ## Lint code
	@echo "Linting code..."
	@golangci-lint run
	@echo "Lint complete!"

# Generate mocks
mocks: ## Generate mocks
	@echo "Generating mocks..."
	@go generate ./...
	@echo "Mocks generated!"

# Run development server with hot reload
dev: ## Run development server with hot reload
	@echo "Starting development server..."
	@air

# Check for security vulnerabilities
security: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	@gosec ./...
	@echo "Security check complete!"

# Build for production
build-prod: ## Build for production
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/evolution-api-go ./cmd/server
	@echo "Production build complete!"

# Run benchmarks
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. ./...
	@echo "Benchmarks complete!"

# Show application info
info: ## Show application info
	@echo "Evolution API Go"
	@echo "Version: 1.0.0"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD)"
	@echo "Build time: $(shell date)"

# Setup development environment
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@cp env.example .env
	@echo "Development environment setup complete!"
	@echo "Please edit .env file with your configuration."
