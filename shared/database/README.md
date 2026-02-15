# Database Package

PostgreSQL connection management and utilities for Nivo services.

## Overview

The `database` package provides a robust PostgreSQL database connection wrapper with connection pooling, health checks, transaction helpers, and context-aware query execution.

## Features

- **Connection Pooling**: Configurable connection pool settings
- **Health Checks**: Built-in database health monitoring
- **Transaction Helpers**: Simplified transaction management with automatic rollback
- **Context Support**: All operations support context for timeouts and cancellation
- **Statistics**: Connection pool statistics for monitoring
- **Graceful Shutdown**: Clean connection closure

## Usage

### Basic Connection

```go
import "github.com/1mb-dev/nivomoney/shared/database"

// Using default configuration
db, err := database.Connect(database.DefaultConfig())
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer db.Close()
```

### Custom Configuration

```go
cfg := database.Config{
    Host:            "db.example.com",
    Port:            5432,
    User:            "myuser",
    Password:        "mypassword",
    Database:        "mydb",
    SSLMode:         "require",
    MaxOpenConns:    25,
    MaxIdleConns:    5,
    ConnMaxLifetime: 5 * time.Minute,
    ConnMaxIdleTime: 1 * time.Minute,
    ConnectTimeout:  5 * time.Second,
}

db, err := database.Connect(cfg)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer db.Close()
```

### From Connection URL

```go
url := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
db, err := database.NewFromURL(url)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
defer db.Close()
```

### Integration with Config Package

```go
import (
    "github.com/1mb-dev/nivomoney/shared/config"
    "github.com/1mb-dev/nivomoney/shared/database"
)

// Load application config
cfg, _ := config.Load()

// Create database config
dbCfg := database.ConfigFromEnv(
    cfg.DatabaseHost,
    cfg.DatabasePort,
    cfg.DatabaseUser,
    cfg.DatabasePassword,
    cfg.DatabaseName,
    cfg.DatabaseSSLMode,
)

// Connect to database
db, err := database.Connect(dbCfg)
```

### Health Checks

```go
ctx := context.Background()

// Simple ping
if err := db.PingContext(ctx); err != nil {
    log.Printf("Database ping failed: %v", err)
}

// Comprehensive health check
if err := db.HealthCheck(ctx); err != nil {
    log.Printf("Database health check failed: %v", err)
}
```

### Queries

```go
ctx := context.Background()

// Query single row
var name string
err := db.QueryRowContext(ctx,
    "SELECT name FROM users WHERE id = $1",
    userID,
).Scan(&name)

// Query multiple rows
rows, err := db.QueryContext(ctx,
    "SELECT id, name FROM users WHERE status = $1",
    "active",
)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        return err
    }
    // Process row
}

// Execute statement
result, err := db.ExecContext(ctx,
    "UPDATE users SET last_login = $1 WHERE id = $2",
    time.Now(),
    userID,
)
```

### Transactions

```go
ctx := context.Background()

// Simple transaction with automatic commit/rollback
err := db.Transaction(ctx, func(tx *sql.Tx) error {
    // Insert user
    var userID int
    err := tx.QueryRowContext(ctx,
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
        "John Doe", "john@example.com",
    ).Scan(&userID)
    if err != nil {
        return err // Transaction will be rolled back
    }

    // Create wallet for user
    _, err = tx.ExecContext(ctx,
        "INSERT INTO wallets (user_id, balance) VALUES ($1, $2)",
        userID, 0,
    )
    if err != nil {
        return err // Transaction will be rolled back
    }

    return nil // Transaction will be committed
})

if err != nil {
    log.Printf("Transaction failed: %v", err)
}
```

### Transaction with Options

```go
// Read-only transaction
opts := &sql.TxOptions{
    ReadOnly: true,
}

err := db.TransactionWithOptions(ctx, opts, func(tx *sql.Tx) error {
    // Perform read operations
    var count int
    return tx.QueryRowContext(ctx,
        "SELECT COUNT(*) FROM users",
    ).Scan(&count)
})

// Serializable isolation level
opts := &sql.TxOptions{
    Isolation: sql.LevelSerializable,
}

err := db.TransactionWithOptions(ctx, opts, func(tx *sql.Tx) error {
    // Perform critical operations
    return nil
})
```

### Connection Pool Statistics

```go
stats := db.Stats()

log.Printf("Open connections: %d", stats.OpenConnections)
log.Printf("In use: %d", stats.InUse)
log.Printf("Idle: %d", stats.Idle)
log.Printf("Wait count: %d", stats.WaitCount)
log.Printf("Wait duration: %v", stats.WaitDuration)
log.Printf("Max idle closed: %d", stats.MaxIdleClosed)
log.Printf("Max lifetime closed: %d", stats.MaxLifetimeClosed)
```

### Graceful Shutdown

```go
// In your main function
db, err := database.Connect(cfg)
if err != nil {
    log.Fatal(err)
}

// Set up signal handling
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// Wait for signal
<-sigChan

// Close database connection
if err := db.Close(); err != nil {
    log.Printf("Error closing database: %v", err)
}
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| Host | Database host | localhost |
| Port | Database port | 5432 |
| User | Database user | nivo |
| Password | Database password | nivo_dev_password |
| Database | Database name | nivo |
| SSLMode | SSL mode (disable, require, verify-ca, verify-full) | disable |
| MaxOpenConns | Maximum open connections | 25 |
| MaxIdleConns | Maximum idle connections | 5 |
| ConnMaxLifetime | Maximum connection lifetime | 5 minutes |
| ConnMaxIdleTime | Maximum connection idle time | 1 minute |
| ConnectTimeout | Connection timeout | 5 seconds |

## Connection Pool Tuning

### Development
```go
cfg := database.DefaultConfig()
// Uses moderate limits suitable for local development
```

### Production
```go
cfg := database.Config{
    Host:            os.Getenv("DB_HOST"),
    Port:            5432,
    MaxOpenConns:    100,  // Higher for production load
    MaxIdleConns:    10,   // Keep more idle connections ready
    ConnMaxLifetime: 10 * time.Minute,
    ConnMaxIdleTime: 5 * time.Minute,
    ConnectTimeout:  10 * time.Second,
}
```

## Best Practices

1. **Always use context**: Pass context to all database operations for timeout control
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   db.QueryRowContext(ctx, query, args...)
   ```

2. **Close rows**: Always close result sets
   ```go
   rows, err := db.QueryContext(ctx, query)
   if err != nil {
       return err
   }
   defer rows.Close()
   ```

3. **Use transactions for related operations**: Group related writes in transactions
   ```go
   db.Transaction(ctx, func(tx *sql.Tx) error {
       // Multiple related operations
   })
   ```

4. **Check errors after rows.Next()**: Check for iteration errors
   ```go
   for rows.Next() {
       // ...
   }
   if err := rows.Err(); err != nil {
       return err
   }
   ```

5. **Use prepared statements for repeated queries**: For performance
   ```go
   stmt, err := db.PrepareContext(ctx, query)
   defer stmt.Close()
   ```

6. **Monitor connection pool**: Track statistics in production
   ```go
   stats := db.Stats()
   metrics.Gauge("db.open_connections", stats.OpenConnections)
   ```

## Testing

The package includes comprehensive tests that require a running PostgreSQL instance:

```bash
# Tests will skip if database is unavailable
go test ./shared/database/...

# Run with coverage
go test -cover ./shared/database/...

# Run against specific database
DB_HOST=testdb.example.com go test ./shared/database/...
```

## Error Handling

```go
import "github.com/1mb-dev/nivomoney/shared/errors"

user, err := getUserFromDB(ctx, db, userID)
if err != nil {
    if err == sql.ErrNoRows {
        return errors.NotFoundWithID("user", userID)
    }
    return errors.DatabaseWrap(err, "failed to fetch user")
}
```

## Related Packages

- [shared/config](../config/README.md) - Configuration management
- [shared/errors](../errors/README.md) - Error handling
- [shared/logger](../logger/README.md) - Logging
