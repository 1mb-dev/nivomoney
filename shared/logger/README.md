# Logger Package

Structured logging for Nivo services using zerolog.

## Overview

The `logger` package provides a structured, performant logging solution built on zerolog. It supports context-aware logging, multiple output formats, and consistent log levels across all services.

## Features

- **Zero-allocation**: Based on zerolog for high performance
- **Structured logging**: JSON and console output formats
- **Context-aware**: Automatic request ID, user ID, correlation ID tracking
- **Multiple levels**: Debug, Info, Warn, Error, Fatal
- **Flexible configuration**: Per-service configuration
- **Global logger**: Optional global instance for convenience

## Usage

### Basic Usage

```go
package main

import (
    "github.com/1mb-dev/nivomoney/shared/logger"
)

func main() {
    // Create logger
    log := logger.NewDefault("my-service")

    // Log messages
    log.Info("service started")
    log.Debug("debug information")
    log.Warn("something might be wrong")
    log.Error("an error occurred")
}
```

### Configuration

```go
log := logger.New(logger.Config{
    Level:       "debug",        // debug, info, warn, error, fatal
    Format:      "json",         // json or console
    ServiceName: "identity-service",
    Output:      os.Stdout,      // optional, defaults to os.Stdout
})
```

### Formatted Logging

```go
log.Infof("user %s logged in", userID)
log.Debugf("processing %d transactions", count)
log.Errorf("failed to connect to %s: %v", host, err)
```

### Structured Fields

```go
// Add single field
log.WithField("user_id", "user-123").
    Info("user action performed")

// Add multiple fields
log.With(map[string]interface{}{
    "transaction_id": "tx-456",
    "amount":         100.50,
    "currency":       "USD",
}).Info("transaction processed")
```

### Context-Aware Logging

```go
ctx := context.Background()
ctx = context.WithValue(ctx, logger.RequestIDKey, "req-abc-123")
ctx = context.WithValue(ctx, logger.UserIDKey, "user-456")

// Logger will automatically include request_id and user_id
contextLog := log.WithContext(ctx)
contextLog.Info("handling request")

// Output: {"level":"info","request_id":"req-abc-123","user_id":"user-456","message":"handling request"}
```

### Error Logging

```go
err := doSomething()
if err != nil {
    log.WithError(err).
        Error("operation failed")
}

// Output includes error in structured format
```

### Output Formats

**Console Format** (Human-readable, colored output for development):
```go
log := logger.New(logger.Config{
    Level:       "info",
    Format:      "console",
    ServiceName: "my-service",
})
```

Output:
```
2:04PM INF service started service=my-service
2:05PM WRN high latency detected latency=250ms service=my-service
2:05PM ERR database connection failed error="connection timeout" service=my-service
```

**JSON Format** (Machine-readable, structured output for production):
```go
log := logger.New(logger.Config{
    Level:       "info",
    Format:      "json",
    ServiceName: "my-service",
})
```

Output:
```json
{"level":"info","service":"my-service","time":"2025-01-15T14:04:05Z","message":"service started"}
{"level":"warn","service":"my-service","latency":250,"time":"2025-01-15T14:05:12Z","message":"high latency detected"}
{"level":"error","service":"my-service","error":"connection timeout","time":"2025-01-15T14:05:15Z","message":"database connection failed"}
```

### Global Logger (Optional)

For convenience, you can use a global logger instance:

```go
// Initialize once at startup
logger.InitGlobal(logger.Config{
    Level:       "info",
    Format:      "console",
    ServiceName: "my-service",
})

// Use anywhere
logger.Info("service started")
logger.Error("something went wrong")
```

### Integration with Services

```go
package main

import (
    "github.com/1mb-dev/nivomoney/shared/config"
    "github.com/1mb-dev/nivomoney/shared/logger"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    // Initialize logger
    log := logger.New(logger.Config{
        Level:       cfg.LogLevel,
        Format:      determineFormat(cfg),
        ServiceName: cfg.ServiceName,
    })

    log.Info("service initialized")

    // Use logger throughout application
    handleRequests(log)
}

func determineFormat(cfg *config.Config) string {
    if cfg.IsDevelopment() {
        return "console"
    }
    return "json"
}
```

### HTTP Middleware Example

```go
func LoggingMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := r.Header.Get("X-Request-ID")
            if requestID == "" {
                requestID = generateRequestID()
            }

            ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

            reqLog := log.WithContext(ctx).WithField("method", r.Method).WithField("path", r.URL.Path)
            reqLog.Info("request received")

            next.ServeHTTP(w, r.WithContext(ctx))

            reqLog.Info("request completed")
        })
    }
}
```

## Log Levels

- **Debug**: Detailed information for diagnosing problems (not shown in production)
- **Info**: General informational messages (service lifecycle, major operations)
- **Warn**: Warning messages (deprecated API usage, poor performance, near errors)
- **Error**: Error messages (operation failures, caught exceptions)
- **Fatal**: Fatal messages that cause program termination

## Context Keys

The package provides standard context keys for common fields:

- `logger.RequestIDKey` - HTTP request ID
- `logger.UserIDKey` - Authenticated user ID
- `logger.CorrelationIDKey` - Distributed tracing correlation ID

## Best Practices

1. **Use Structured Fields**: Prefer structured fields over formatted strings
   ```go
   // Good
   log.WithField("user_id", userID).Info("user logged in")

   // Less ideal
   log.Infof("user %s logged in", userID)
   ```

2. **Log at Appropriate Levels**:
   - Debug: Development debugging only
   - Info: Important service events
   - Warn: Recoverable issues
   - Error: Operation failures
   - Fatal: Critical failures (use sparingly)

3. **Include Context**: Always propagate context for request tracking
   ```go
   contextLog := log.WithContext(ctx)
   ```

4. **Avoid Logging Secrets**: Never log passwords, tokens, or sensitive data

5. **Use JSON in Production**: JSON format is easier to parse and analyze

## Performance

Zerolog is one of the fastest Go logging libraries:
- Zero allocation in hot paths
- Minimal CPU overhead
- Efficient JSON encoding

## Testing

```bash
go test ./shared/logger/...
go test -cover ./shared/logger/...
```

## Related Packages

- [shared/config](../config/README.md) - Configuration management
- [shared/middleware](../middleware/README.md) - HTTP middleware (to be created)
