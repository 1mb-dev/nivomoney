# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN GOTOOLCHAIN=auto go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo -o bin/identity-service ./services/identity/cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/identity-service .

# Copy migrations
COPY --from=builder /app/services/identity/migrations ./migrations

# Expose port
EXPOSE 8080

# Run the service
CMD ["./identity-service"]
