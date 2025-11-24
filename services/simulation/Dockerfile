# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY services/simulation/go.mod services/simulation/go.sum* ./services/simulation/
COPY shared/go.mod shared/go.sum* ./shared/

# Download dependencies
WORKDIR /build/services/simulation
RUN go mod download

# Copy source code
WORKDIR /build
COPY services/simulation/ ./services/simulation/
COPY shared/ ./shared/

# Build the application
WORKDIR /build/services/simulation
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o simulation ./cmd/server

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/services/simulation/simulation .

# Expose port (simulation runs as background service)
EXPOSE 8086

# Run the application
CMD ["./simulation"]
