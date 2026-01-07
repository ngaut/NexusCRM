package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexuscrm/backend/pkg/utils"
	"github.com/nexuscrm/shared/pkg/constants"
)

// OutboxEvent represents a persisted event record
type OutboxEvent struct {
	ID               string
	EventType        string
	Payload          string
	Status           string
	RetryCount       int
	ErrorMessage     string
	CreatedDate      time.Time
	ProcessedDate    sql.NullTime
	LastModifiedDate time.Time
}

// OutboxRepository handles database operations for the outbox pattern
type OutboxRepository struct {
	db *sql.DB
}

// NewOutboxRepository creates a new OutboxRepository
func NewOutboxRepository(db *sql.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

// Enqueue inserts a new event into the outbox
func (r *OutboxRepository) Enqueue(ctx context.Context, exec Executor, eventType string, payload interface{}) (string, error) {
	id := utils.GenerateID()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, event_type, payload, status, retry_count, created_date, last_modified_date)
		VALUES (?, ?, ?, ?, 0, NOW(), NOW())
	`, constants.TableOutboxEvent)

	// Status defaults to 'pending' from constant (imported or defined locally? Using string literal or passing it in?)
	// OutboxService defined constants. Ideally constants should be in shared.
	// For now, I'll accept 'status' or just hardcode 'pending' if this is Enqueue.
	// OutboxService uses "pending".
	status := "pending"

	_, err = exec.ExecContext(ctx, query, id, eventType, payloadJSON, status)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue event: %w", err)
	}

	return id, nil
}

// GetPendingEvents retrieves IDs of pending events ordered by creation time
func (r *OutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := fmt.Sprintf(`
		SELECT id, event_type, payload, retry_count
		FROM %s
		WHERE status = ?
		ORDER BY created_date ASC
		LIMIT ?
	`, constants.TableOutboxEvent)

	rows, err := r.db.QueryContext(ctx, query, "pending", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.RetryCount); err != nil {
			continue
		}
		events = append(events, e)
	}

	return events, nil
}

// ClaimEvent attempts to lock a specific event for processing
func (r *OutboxRepository) ClaimEvent(ctx context.Context, exec Executor, id string) (string, error) {
	query := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE id = ? AND status = ? 
		FOR UPDATE SKIP LOCKED
	`, constants.TableOutboxEvent)

	var claimedID string
	err := exec.QueryRowContext(ctx, query, id, "pending").Scan(&claimedID)
	if err == sql.ErrNoRows {
		return "", nil // Already claimed
	}
	if err != nil {
		return "", err
	}
	return claimedID, nil
}

// UpdateStatus updates the status and related fields of an event
func (r *OutboxRepository) UpdateStatus(ctx context.Context, exec Executor, id string, status string, errMessage string) error {
	var query string
	var args []interface{}

	if status == "processed" {
		query = fmt.Sprintf(`
			UPDATE %s 
			SET status = ?, processed_date = NOW(), last_modified_date = NOW()
			WHERE id = ?
		`, constants.TableOutboxEvent)
		args = []interface{}{status, id}
	} else if status == "failed" {
		query = fmt.Sprintf(`
			UPDATE %s 
			SET status = ?, error_message = ?, last_modified_date = NOW()
			WHERE id = ?
		`, constants.TableOutboxEvent)
		args = []interface{}{status, errMessage, id}
	} else {
		return fmt.Errorf("unsupported status update: %s", status)
	}

	_, err := exec.ExecContext(ctx, query, args...)
	return err
}

// IncrementRetry increments the retry count and updates error message
func (r *OutboxRepository) IncrementRetry(ctx context.Context, exec Executor, id string, newCount int, errMessage string) error {
	query := fmt.Sprintf(`
		UPDATE %s 
		SET retry_count = ?, error_message = ?, last_modified_date = NOW()
		WHERE id = ?
	`, constants.TableOutboxEvent)

	_, err := exec.ExecContext(ctx, query, newCount, errMessage, id)
	return err
}

// CleanupProcessed deletes old processed events
func (r *OutboxRepository) CleanupProcessed(ctx context.Context, cutoff time.Time) (int64, error) {
	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE status = ? AND processed_date < ?
	`, constants.TableOutboxEvent)

	result, err := r.db.ExecContext(ctx, query, "processed", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
