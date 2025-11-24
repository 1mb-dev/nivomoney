package database

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 5432 {
		t.Errorf("Expected port 5432, got %d", cfg.Port)
	}
	if cfg.MaxOpenConns != 25 {
		t.Errorf("Expected MaxOpenConns 25, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 5 {
		t.Errorf("Expected MaxIdleConns 5, got %d", cfg.MaxIdleConns)
	}
}

func TestConfigFromEnv(t *testing.T) {
	cfg := ConfigFromEnv(
		"db.example.com",
		5433,
		"testuser",
		"testpass",
		"testdb",
		"require",
	)

	if cfg.Host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", cfg.Host)
	}
	if cfg.Port != 5433 {
		t.Errorf("Expected port 5433, got %d", cfg.Port)
	}
	if cfg.User != "testuser" {
		t.Errorf("Expected user 'testuser', got '%s'", cfg.User)
	}
	if cfg.Database != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", cfg.Database)
	}
	if cfg.SSLMode != "require" {
		t.Errorf("Expected sslmode 'require', got '%s'", cfg.SSLMode)
	}
}

// Note: The following tests require a running PostgreSQL instance.
// They will be skipped if the database is not available.

func getTestDB(t *testing.T) *DB {
	t.Helper()

	cfg := DefaultConfig()
	cfg.ConnectTimeout = 2 * time.Second

	db, err := Connect(cfg)
	if err != nil {
		t.Skipf("Skipping test: database not available: %v", err)
		return nil
	}

	return db
}

func TestConnect(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	// Verify connection is working
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Errorf("Failed to ping database: %v", err)
	}
}

func TestHealthCheck(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := db.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}

func TestStats(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	stats := db.Stats()

	// Just verify we can get stats
	if stats.MaxOpenConnections < 0 {
		t.Error("Invalid stats returned")
	}
}

func TestTransaction(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a test table
	_, err := db.ExecContext(ctx, `
		CREATE TEMP TABLE test_transaction (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test successful transaction
	err = db.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO test_transaction (value) VALUES ($1)", "test1")
		return err
	})
	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Verify data was inserted
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_transaction").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row, got %d", count)
	}

	// Test transaction rollback on error
	err = db.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "INSERT INTO test_transaction (value) VALUES ($1)", "test2")
		if err != nil {
			return err
		}
		// Force an error to trigger rollback
		return sql.ErrTxDone
	})
	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback - should still have only 1 row
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_transaction").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row after rollback, got %d", count)
	}
}

func TestTransactionWithOptions(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create a test table
	_, err := db.ExecContext(ctx, `
		CREATE TEMP TABLE test_transaction_opts (
			id SERIAL PRIMARY KEY,
			value TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Test with read-only transaction
	opts := &sql.TxOptions{
		ReadOnly: true,
	}

	err = db.TransactionWithOptions(ctx, opts, func(tx *sql.Tx) error {
		var count int
		return tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM test_transaction_opts").Scan(&count)
	})
	if err != nil {
		t.Errorf("Read-only transaction failed: %v", err)
	}
}

func TestQueryRowContext(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	var result int
	err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("QueryRowContext failed: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected result 1, got %d", result)
	}
}

func TestQueryContext(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	rows, err := db.QueryContext(ctx, "SELECT 1 UNION SELECT 2")
	if err != nil {
		t.Fatalf("QueryContext failed: %v", err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestExecContext(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Create and drop a temp table
	_, err := db.ExecContext(ctx, "CREATE TEMP TABLE test_exec (id INT)")
	if err != nil {
		t.Errorf("ExecContext failed: %v", err)
	}
}

func TestConnect_InvalidConfig(t *testing.T) {
	cfg := Config{
		Host:           "invalid-host-that-does-not-exist.example.com",
		Port:           5432,
		User:           "test",
		Password:       "test",
		Database:       "test",
		SSLMode:        "disable",
		ConnectTimeout: 1 * time.Second,
	}

	_, err := Connect(cfg)
	if err == nil {
		t.Error("Expected Connect to fail with invalid config")
	}
}

func TestNewFromURL(t *testing.T) {
	// Test with invalid URL
	_, err := NewFromURL("invalid-url")
	if err == nil {
		t.Error("Expected NewFromURL to fail with invalid URL")
	}

	// Test with valid URL but unreachable host
	url := "postgres://user:pass@invalid-host:5432/db?connect_timeout=1"
	_, err = NewFromURL(url)
	if err == nil {
		t.Error("Expected NewFromURL to fail with unreachable host")
	}
}

func TestMustConnect(t *testing.T) {
	// Test that MustConnect panics on error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustConnect to panic with invalid config")
		}
	}()

	cfg := Config{
		Host:           "invalid-host",
		Port:           5432,
		User:           "test",
		Password:       "test",
		Database:       "test",
		SSLMode:        "disable",
		ConnectTimeout: 1 * time.Second,
	}

	MustConnect(cfg)
}
