package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

// getExecutor returns the provided executor or defaults to internal DB
func (r *OutboxRepository) getExecutor(exec Executor) Executor {
	if exec != nil {
		return exec
	}
	return r.db
}

// Enqueue inserts a new event into the outbox
func (r *OutboxRepository) Enqueue(ctx context.Context, exec Executor, eventType string, payload interface{}) (string, error) {
	executor := r.getExecutor(exec)
	id := utils.GenerateID()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event payload: %w", err)
	}

	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysOutboxEvent_EventType, constants.FieldSysOutboxEvent_Payload,
		constants.FieldSysOutboxEvent_Status, constants.FieldSysOutboxEvent_RetryCount,
		constants.FieldCreatedDate, constants.FieldLastModifiedDate,
	}, ", ")

	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, 0, NOW(), NOW())
	`, constants.TableOutboxEvent, cols)

	// Status defaults to 'pending'
	// Ideally constants should be in shared (currently used from constants package)
	status := constants.OutboxStatusPending

	_, err = executor.ExecContext(ctx, query, id, eventType, payloadJSON, status)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue event: %w", err)
	}

	return id, nil
}

// GetPendingEvents retrieves IDs of pending events ordered by creation time
func (r *OutboxRepository) GetPendingEvents(ctx context.Context, limit int) ([]OutboxEvent, error) {
	cols := strings.Join([]string{
		constants.FieldID, constants.FieldSysOutboxEvent_EventType, constants.FieldSysOutboxEvent_Payload, constants.FieldSysOutboxEvent_RetryCount,
	}, ", ")

	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE %s = ?
		ORDER BY %s ASC
		LIMIT ?
	`, cols, constants.TableOutboxEvent, constants.FieldSysOutboxEvent_Status, constants.FieldCreatedDate)

	rows, err := r.db.QueryContext(ctx, query, constants.OutboxStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.RetryCount); err != nil {
			log.Printf("Warning: failed to scan outbox event: %v", err)
			continue
		}
		events = append(events, e)
	}

	return events, nil
}

// ClaimEvent attempts to lock a specific event for processing
func (r *OutboxRepository) ClaimEvent(ctx context.Context, exec Executor, id string) (string, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE %s = ? AND %s = ? 
		FOR UPDATE SKIP LOCKED
	`, constants.FieldID, constants.TableOutboxEvent, constants.FieldID, constants.FieldSysOutboxEvent_Status)

	var claimedID string
	err := exec.QueryRowContext(ctx, query, id, constants.OutboxStatusPending).Scan(&claimedID)
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

	if status == string(constants.OutboxStatusProcessed) {
		query = fmt.Sprintf(`
			UPDATE %s 
			SET %s = ?, %s = NOW(), %s = NOW()
			WHERE %s = ?
		`, constants.TableOutboxEvent, constants.FieldSysOutboxEvent_Status, constants.FieldSysOutboxEvent_ProcessedDate, constants.FieldLastModifiedDate, constants.FieldID)
		args = []interface{}{status, id}
	} else if status == string(constants.OutboxStatusFailed) {
		query = fmt.Sprintf(`
			UPDATE %s 
			SET %s = ?, %s = ?, %s = NOW()
			WHERE %s = ?
		`, constants.TableOutboxEvent, constants.FieldSysOutboxEvent_Status, constants.FieldSysOutboxEvent_ErrorMessage, constants.FieldLastModifiedDate, constants.FieldID)
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
		SET %s = ?, %s = ?, %s = NOW()
		WHERE %s = ?
	`, constants.TableOutboxEvent, constants.FieldSysOutboxEvent_RetryCount, constants.FieldSysOutboxEvent_ErrorMessage, constants.FieldLastModifiedDate, constants.FieldID)

	_, err := exec.ExecContext(ctx, query, newCount, errMessage, id)
	return err
}

// CleanupProcessed deletes old processed events
func (r *OutboxRepository) CleanupProcessed(ctx context.Context, cutoff time.Time) (int64, error) {
	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE %s = ? AND %s < ?
	`, constants.TableOutboxEvent, constants.FieldSysOutboxEvent_Status, constants.FieldSysOutboxEvent_ProcessedDate)

	result, err := r.db.ExecContext(ctx, query, constants.OutboxStatusProcessed, cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
