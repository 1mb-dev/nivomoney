# Config Package

Centralized configuration management for Nivo services.

## Overview

The `config` package provides type-safe configuration loading from environment variables with sensible defaults. All services use this package for consistent configuration management.

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "github.com/vnykmshr/nivo/shared/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Use configuration
    log.Printf("Starting %s on port %d", cfg.ServiceName, cfg.ServicePort)
    log.Printf("Environment: %s", cfg.Environment)
    log.Printf("Database: %s", cfg.DatabaseURL)
}
```

### Environment Variables

All configuration can be overridden via environment variables:

#### Application
- `ENVIRONMENT` - Environment (development/production) [default: "development"]
- `SERVICE_NAME` - Service name [default: "nivo"]
- `SERVICE_PORT` - Service port [default: 8080]
- `LOG_LEVEL` - Log level (debug/info/warn/error) [default: "info"]

#### Database
- `DATABASE_URL` - Complete database connection string
- `DATABASE_HOST` - Database host [default: "localhost"]
- `DATABASE_PORT` - Database port [default: 5432]
- `DATABASE_USER` - Database user [default: "nivo"]
- `DATABASE_PASSWORD` - Database password [default: "nivo_dev_password"]
- `DATABASE_NAME` - Database name [default: "nivo"]
- `DATABASE_SSL_MODE` - SSL mode [default: "disable"]

#### Redis
- `REDIS_URL` - Complete Redis connection string
- `REDIS_HOST` - Redis host [default: "localhost"]
- `REDIS_PORT` - Redis port [default: 6379]
- `REDIS_PASSWORD` - Redis password [default: "nivo_redis_password"]
- `REDIS_DB` - Redis database number [default: 0]

#### NSQ
- `NSQLOOKUPD_ADDR` - NSQ Lookup daemon address [default: "localhost:4161"]
- `NSQD_ADDR` - NSQ daemon address [default: "localhost:4150"]

#### JWT
- `JWT_SECRET` - JWT signing secret [default: "nivo-dev-secret-change-in-production"]
- `JWT_EXPIRY` - JWT token expiry duration [default: "24h"]
- `JWT_REFRESH_EXPIRY` - JWT refresh token expiry [default: "168h"]

#### Server
- `SERVER_READ_TIMEOUT` - HTTP read timeout [default: "10s"]
- `SERVER_WRITE_TIMEOUT` - HTTP write timeout [default: "10s"]
- `SERVER_IDLE_TIMEOUT` - HTTP idle timeout [default: "120s"]

#### Observability
- `PROMETHEUS_PORT` - Prometheus metrics port [default: 9090]
- `ENABLE_PROFILING` - Enable pprof profiling [default: false]

### Validation

Configuration is automatically validated when loaded in production:

```go
cfg, err := config.Load()
if err != nil {
    // Handle validation errors
    log.Fatalf("Invalid configuration: %v", err)
}
```

Production validation checks:
- JWT secret must not be the default value
- Database password must not be the default value
- Redis password must not be the default value
- Ports must be in valid range (1-65535)

### Environment Detection

```go
cfg, _ := config.Load()

if cfg.IsDevelopment() {
    log.Println("Running in development mode")
}

if cfg.IsProduction() {
    log.Println("Running in production mode")
}
```

## Example .env File

```bash
# Application
ENVIRONMENT=development
SERVICE_NAME=identity-service
SERVICE_PORT=8081
LOG_LEVEL=debug

# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=nivo
DATABASE_PASSWORD=secure_password
DATABASE_NAME=nivo

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# JWT
JWT_SECRET=my-secret-key
JWT_EXPIRY=24h
```

## Testing

The package includes comprehensive tests:

```bash
go test ./shared/config/...
go test -cover ./shared/config/...
```

## Design Principles

1. **Sensible Defaults**: Works out of the box for development
2. **Environment-First**: All config via environment variables
3. **Type Safety**: Strong typing for all configuration values
4. **Validation**: Automatic validation in production
5. **Zero Dependencies**: Uses only standard library
