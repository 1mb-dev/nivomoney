# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN GOTOOLCHAIN=auto go mod download

# Copy source code
COPY shared/ ./shared/
COPY services/risk/ ./services/risk/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -a -installsuffix cgo -o bin/risk-service ./services/risk/cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/risk-service .

# Copy migrations
COPY --from=builder /app/services/risk/migrations ./migrations

# Expose port
EXPOSE 8085

# Run the service
CMD ["./risk-service"]
