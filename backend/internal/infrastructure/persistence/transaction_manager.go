package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
)

// txContextKey is the key for storing transaction in context
type txContextKey struct{}

// TransactionManager handles database transactions with retry logic for deadlocks
type TransactionManager struct {
	db *database.TiDBConnection
}

// NewTransactionManager creates a new TransactionManager
func NewTransactionManager(db *database.TiDBConnection) *TransactionManager {
	return &TransactionManager{db: db}
}

// WithTransaction executes a function within a database transaction.
// The transaction is automatically rolled back if the function returns an error or panics.
// The transaction is committed if the function returns nil.
func (tm *TransactionManager) WithTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := tm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed: %w (rollback error: %v)", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithRetry executes a function within a transaction with automatic retry on deadlock.
// Deadlocks are retried up to maxRetries times with exponential backoff.
// Other errors are returned immediately without retry.
func (tm *TransactionManager) WithRetry(fn func(tx *sql.Tx) error, maxRetries int) error {
	if maxRetries < 1 {
		maxRetries = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := tm.WithTransaction(fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if it's a deadlock error
		if !isDeadlock(err) {
			return err // Not a deadlock, return immediately
		}

		// Deadlock detected, retry with exponential backoff
		if attempt < maxRetries-1 {
			backoff := time.Millisecond * time.Duration(100*(1<<uint(attempt)))
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

// InjectTx injects a transaction into the context
func (tm *TransactionManager) InjectTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// ExtractTx extracts a transaction from the context
func (tm *TransactionManager) ExtractTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// WithIsolationLevel executes a function within a transaction with a specific isolation level.
// Supported levels: READ UNCOMMITTED, READ COMMITTED, REPEATABLE READ, SERIALIZABLE
func (tm *TransactionManager) WithIsolationLevel(
	level IsolationLevel,
	fn func(tx *sql.Tx) error,
) error {
	return tm.WithTransaction(func(tx *sql.Tx) error {
		// Set isolation level
		_, err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level))
		if err != nil {
			return fmt.Errorf("failed to set isolation level: %w", err)
		}

		// Execute the function
		return fn(tx)
	})
}

// IsolationLevel represents SQL transaction isolation levels
type IsolationLevel string

const (
	ReadUncommitted IsolationLevel = "READ UNCOMMITTED"
	ReadCommitted   IsolationLevel = "READ COMMITTED"
	RepeatableRead  IsolationLevel = "REPEATABLE READ"
	Serializable    IsolationLevel = "SERIALIZABLE"
)

// isDeadlock checks if an error is a deadlock error.
// MySQL/TiDB deadlock error codes:
// - 1213: Deadlock found when trying to get lock
// - 1205: Lock wait timeout exceeded
func isDeadlock(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "deadlock") ||
		strings.Contains(errMsg, "lock wait timeout") ||
		strings.Contains(errMsg, "1213") ||
		strings.Contains(errMsg, "1205")
}
