package database

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

// TiDBConnection represents a TiDB database connection
// Note: sql.DB is already thread-safe and manages its own connection pool.
// We do NOT wrap it with additional mutexes as that causes deadlocks under
// high concurrency (writers waiting for connections block readers).
type TiDBConnection struct {
	db *sql.DB
}

var (
	instance *TiDBConnection
	once     sync.Once
	initErr  error
	tlsOnce  sync.Once // Ensure TLS config is registered only once
)

// GetInstance returns the singleton TiDB connection
func GetInstance() (*TiDBConnection, error) {
	once.Do(func() {
		instance, initErr = newConnection()
	})
	return instance, initErr
}

// newConnection creates a new TiDB connection
func newConnection() (*TiDBConnection, error) {
	host := os.Getenv("TIDB_HOST")
	port := os.Getenv("TIDB_PORT")
	user := os.Getenv("TIDB_USER")
	password := os.Getenv("TIDB_PASSWORD")
	database := os.Getenv("TIDB_DATABASE")

	if port == "" {
		port = "4000"
	}

	if database == "" {
		database = "nexuscrm"
	}

	// Determine TLS configuration based on host
	tlsParam := ""
	if host != "" && host != "127.0.0.1" && host != "localhost" {
		// Remote host (e.g., TiDB Cloud) - register TLS config with ServerName
		// Use sync.Once to prevent panic on duplicate registration (e.g., in tests)
		tlsOnce.Do(func() {
			if err := mysql.RegisterTLSConfig("tidb", &tls.Config{
				MinVersion: tls.VersionTLS12,
				ServerName: host, // Required for TLS verification
			}); err != nil {
				// Just log as we can't return error from sync.Once
				log.Printf("Failed to register TLS config: %v\n", err)
			}
		})
		tlsParam = "&tls=tidb"
	}
	// For localhost, no TLS is used

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
		user, password, host, port, database, tlsParam)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	// IMPORTANT: MaxIdleConns must equal MaxOpenConns to prevent port exhaustion.
	// If MaxIdleConns < MaxOpenConns, connections are closed/reopened frequently,
	// which exhausts ephemeral ports under high concurrency.
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(100) // Match MaxOpenConns to keep connections alive

	// Connection lifecycle settings for auto-reconnection
	// MaxLifetime ensures connections are recycled before they become stale
	db.SetConnMaxLifetime(5 * time.Minute)
	// MaxIdleTime closes idle connections that haven't been used recently
	db.SetConnMaxIdleTime(3 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &TiDBConnection{db: db}, nil
}

// Query executes a SELECT query and returns rows
// sql.DB handles connection pooling and concurrency internally
func (c *TiDBConnection) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.Query(query, args...)
}

// QueryContext executes a SELECT query with context
func (c *TiDBConnection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a SELECT query that returns at most one row
func (c *TiDBConnection) QueryRow(query string, args ...interface{}) *sql.Row {
	return c.db.QueryRow(query, args...)
}

// QueryRowContext executes a SELECT query with context that returns at most one row
func (c *TiDBConnection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

// Exec executes an INSERT, UPDATE, or DELETE query
func (c *TiDBConnection) Exec(query string, args ...interface{}) (sql.Result, error) {
	return c.db.Exec(query, args...)
}

// ExecContext executes an INSERT, UPDATE, or DELETE query with context
func (c *TiDBConnection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Begin starts a new transaction
func (c *TiDBConnection) Begin() (*sql.Tx, error) {
	return c.db.Begin()
}

// BeginTx starts a new transaction with context
func (c *TiDBConnection) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

// DB returns the underlying *sql.DB connection
// This is useful for operations that need direct access to sql.DB
func (c *TiDBConnection) DB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *TiDBConnection) Close() error {
	return c.db.Close()
}
