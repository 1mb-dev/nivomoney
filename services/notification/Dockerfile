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
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo -o bin/notification-service ./services/notification/cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/notification-service .

# Copy migrations
COPY --from=builder /app/services/notification/migrations ./migrations

# Expose port
EXPOSE 8087

# Run the service
CMD ["./notification-service"]
