.PHONY: help build build-all test test-coverage test-integration lint fmt vet clean clean-all docker-up docker-down docker-logs run migrate-up migrate-down

# Default target - show help
help:
	@echo "Nivo - Development Commands"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              Build all services"
	@echo "  make build-all          Build all services and gateway"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               Run all tests"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make test-integration   Run integration tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint               Run linters (requires golangci-lint)"
	@echo "  make fmt                Format code with gofmt"
	@echo "  make vet                Run go vet"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-up          Start all containers (postgres, redis, etc.)"
	@echo "  make docker-down        Stop all containers"
	@echo "  make docker-logs        Show container logs"
	@echo ""
	@echo "Development:"
	@echo "  make run                Run all services locally"
	@echo "  make migrate-up         Run database migrations up"
	@echo "  make migrate-down       Rollback database migrations"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean              Remove build artifacts"
	@echo "  make clean-all          Remove all generated files (including vendor/)"
	@echo ""

# Build targets
build:
	@echo "Building all services..."
	@go build -o bin/gateway ./gateway
	@echo "Gateway built successfully"
	@for service in services/*; do \
		if [ -d "$$service" ] && [ -f "$$service/main.go" ]; then \
			echo "Building $$service..."; \
			go build -o bin/$$(basename $$service) ./$$service; \
		fi \
	done
	@echo "Build complete!"

build-all: build
	@echo "All services built"

# Test targets
test:
	@echo "Running tests..."
	@go test -v -race ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	@echo "Running integration tests..."
	@go test -v -race -tags=integration ./...

# Code quality targets
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: make install-lint"; \
		exit 1; \
	fi

fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Code formatted"

vet:
	@echo "Running go vet..."
	@go vet ./...

# Docker targets
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Containers started"

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "Containers stopped"

docker-logs:
	@docker-compose logs -f

# Development targets
run:
	@echo "Starting services..."
	@echo "Note: Implement service runners as services are developed"

migrate-up:
	@echo "Running migrations..."
	@echo "Note: Migration tool to be configured"

migrate-down:
	@echo "Rolling back migrations..."
	@echo "Note: Migration tool to be configured"

# Cleanup targets
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

clean-all: clean
	@echo "Removing all generated files..."
	@rm -rf vendor/
	@go clean -cache -testcache -modcache
	@echo "Deep clean complete"

# Development tooling installation helpers
install-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

install-tools: install-lint
	@echo "All development tools installed"
