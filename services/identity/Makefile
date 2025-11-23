.PHONY: help build run test clean migrate-up migrate-down docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the Identity Service binary"
	@echo "  run          - Run the Identity Service"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

# Build the service
build:
	@echo "Building Identity Service..."
	@cd ../../ && go build -o bin/identity-service ./services/identity/cmd/server

# Run the service
run:
	@echo "Running Identity Service..."
	@cd ../../ && go run ./services/identity/cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@cd ../../ && go test -v -cover ./services/identity/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@cd ../../ && rm -f bin/identity-service

# Run migrations up
migrate-up:
	@echo "Running migrations..."
	@cd ../../ && go run ./services/identity/cmd/server/main.go --migrate-only

# Run migrations down (not implemented for safety)
migrate-down:
	@echo "Down migrations not implemented - use manual rollback for safety"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@cd ../../ && docker build -t nivo/identity-service:latest -f services/identity/Dockerfile .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file services/identity/.env nivo/identity-service:latest

# Development helpers
.PHONY: dev
dev:
	@echo "Starting development server with auto-reload..."
	@cd ../../ && air -c services/identity/.air.toml
