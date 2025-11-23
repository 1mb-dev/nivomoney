.PHONY: help build run test clean dev docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the gateway binary
	@echo "Building gateway..."
	@go build -o bin/gateway ./cmd/server

run: ## Run the gateway locally
	@echo "Running gateway..."
	@go run ./cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

dev: ## Run in development mode with air (hot reload)
	@echo "Running in development mode..."
	@air -c .air.toml || go run ./cmd/server/main.go

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t nivo-gateway:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8000:8000 --env-file .env nivo-gateway:latest

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run
