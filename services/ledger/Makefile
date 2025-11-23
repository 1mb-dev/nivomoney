.PHONY: help build run test clean

# Default target
help:
	@echo "Available targets:"
	@echo "  build  - Build the Ledger Service binary"
	@echo "  run    - Run the Ledger Service"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts"

# Build the service
build:
	@echo "Building Ledger Service..."
	@cd ../../ && go build -o bin/ledger-service ./services/ledger/cmd/server

# Run the service
run:
	@echo "Running Ledger Service..."
	@cd ../../ && go run ./services/ledger/cmd/server/main.go

# Run tests
test:
	@echo "Running tests..."
	@cd ../../ && go test -v -cover ./services/ledger/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@cd ../../ && rm -f bin/ledger-service
