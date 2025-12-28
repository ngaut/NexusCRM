package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
)

// OutboxEvent status constants
const (
	OutboxStatusPending   = "pending"
	OutboxStatusProcessed = "processed"
	OutboxStatusFailed    = "failed"
	MaxRetryAttempts      = 5
)

// OutboxService handles transactional event storage and async publishing.
// It implements the Outbox Pattern for guaranteed event delivery.
type OutboxService struct {
	db        *database.TiDBConnection
	eventBus  *EventBus
	txManager *TransactionManager

	// Worker control
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewOutboxService creates a new OutboxService
func NewOutboxService(db *database.TiDBConnection, eventBus *EventBus, txManager *TransactionManager) *OutboxService {
	return &OutboxService{
		db:        db,
		eventBus:  eventBus,
		txManager: txManager,
		stopCh:    make(chan struct{}),
	}
}

// EnqueueEvent stores an event in the outbox table within the current transaction.
// This ensures the event is persisted atomically with the business operation.
func (os *OutboxService) EnqueueEvent(ctx context.Context, eventType events.EventType, payload RecordEventPayload) error {
	// Try to extract existing transaction from context
	tx := os.txManager.ExtractTx(ctx)
	if tx != nil {
		return os.enqueueWithTx(ctx, tx, eventType, payload)
	}

	// No transaction in context, execute directly
	return os.enqueueDirect(ctx, eventType, payload)
}

// EnqueueEventTx stores an event using an explicit transaction
func (os *OutboxService) EnqueueEventTx(ctx context.Context, tx *sql.Tx, eventType events.EventType, payload RecordEventPayload) error {
	return os.enqueueWithTx(ctx, tx, eventType, payload)
}

// enqueueWithTx inserts event into outbox using the provided transaction
func (os *OutboxService) enqueueWithTx(ctx context.Context, tx *sql.Tx, eventType events.EventType, payload RecordEventPayload) error {
	id := GenerateID()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, event_type, payload, status, retry_count, created_date)
		VALUES (?, ?, ?, ?, 0, NOW())
	`, constants.TableOutboxEvent)

	_, err = tx.ExecContext(ctx, query, id, string(eventType), payloadJSON, OutboxStatusPending)
	if err != nil {
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	log.Printf("✅ [Outbox] Enqueued event %s (Type: %s, ID: %s)", eventType, string(eventType), id)
	return nil
}

// enqueueDirect inserts event directly without transaction context
func (os *OutboxService) enqueueDirect(ctx context.Context, eventType events.EventType, payload RecordEventPayload) error {
	id := GenerateID()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, event_type, payload, status, retry_count, created_date)
		VALUES (?, ?, ?, ?, 0, NOW())
	`, constants.TableOutboxEvent)

	_, err = os.db.DB().ExecContext(ctx, query, id, string(eventType), payloadJSON, OutboxStatusPending)
	if err != nil {
		return fmt.Errorf("failed to enqueue event: %w", err)
	}

	log.Printf("✅ [Outbox] Enqueued event %s (Type: %s, ID: %s)", eventType, string(eventType), id)
	return nil
}

// StartWorker starts the background worker that processes pending outbox events.
// The worker polls with the specified interval.
func (os *OutboxService) StartWorker(interval time.Duration) {
	os.wg.Add(1)
	go func() {
		defer os.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("📤 Outbox worker started with %v interval", interval)

		for {
			select {
			case <-os.stopCh:
				log.Printf("📤 Outbox worker stopping...")
				return
			case <-ticker.C:
				if err := os.ProcessOutbox(context.Background()); err != nil {
					log.Printf("⚠️ Outbox worker error: %v", err)
				}
			}
		}
	}()
}

// StopWorker stops the background worker gracefully
func (os *OutboxService) StopWorker() {
	os.stopOnce.Do(func() {
		close(os.stopCh)
	})
	os.wg.Wait()
	log.Printf("📤 Outbox worker stopped")
}

// ProcessOutbox processes all pending events in the outbox table.
// Events are published via EventBus and marked as processed.
// Each event is processed in its own transaction to ensure atomicity.
func (os *OutboxService) ProcessOutbox(ctx context.Context) error {
	// First, get IDs of pending events (non-locking query to minimize contention)
	query := fmt.Sprintf(`
		SELECT id, event_type, payload, retry_count
		FROM %s
		WHERE status = ?
		ORDER BY created_date ASC
		LIMIT 100
	`, constants.TableOutboxEvent)

	rows, err := os.db.DB().QueryContext(ctx, query, OutboxStatusPending)
	if err != nil {
		return fmt.Errorf("failed to query pending events: %w", err)
	}
	defer rows.Close()

	type pendingEvent struct {
		ID         string
		EventType  string
		Payload    string
		RetryCount int
	}

	var eventsToProcess []pendingEvent
	for rows.Next() {
		var e pendingEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.RetryCount); err != nil {
			log.Printf("⚠️ Failed to scan outbox event: %v", err)
			continue
		}
		eventsToProcess = append(eventsToProcess, e)
	}

	if len(eventsToProcess) > 0 {
		log.Printf("🔄 [Outbox] Processing %d pending events", len(eventsToProcess))
	}

	// Check for errors during row iteration
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating pending events: %w", err)
	}

	// Process each event atomically (claim, process, update status)
	for _, e := range eventsToProcess {
		if err := os.processEventAtomic(ctx, e.ID, e.EventType, e.Payload, e.RetryCount); err != nil {
			log.Printf("⚠️ Failed to process outbox event %s: %v", e.ID, err)
		}
	}

	return nil
}

// processEventAtomic claims an event, publishes it, and updates status atomically
func (os *OutboxService) processEventAtomic(ctx context.Context, id, eventType, payloadJSON string, retryCount int) error {
	// Start transaction to claim the event
	tx, err := os.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Try to claim this specific event (skip if already claimed by another worker)
	claimQuery := fmt.Sprintf(`
		SELECT id FROM %s 
		WHERE id = ? AND status = ? 
		FOR UPDATE SKIP LOCKED
	`, constants.TableOutboxEvent)

	var claimedID string
	if err := tx.QueryRowContext(ctx, claimQuery, id, OutboxStatusPending).Scan(&claimedID); err != nil {
		if err == sql.ErrNoRows {
			// Already processed by another worker, skip
			return nil
		}
		return fmt.Errorf("failed to claim event: %w", err)
	}

	// Parse payload
	var payload RecordEventPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		// Mark as failed within transaction
		log.Printf("❌ [Outbox] Event %s failed payload unmarshal: %v", id, err)
		if markErr := os.markFailedTx(ctx, tx, id, fmt.Sprintf("invalid payload: %v", err)); markErr != nil {
			return fmt.Errorf("failed to mark event as failed: %w", markErr)
		}
		return tx.Commit()
	}

	// Publish via EventBus
	if err := os.eventBus.Publish(ctx, events.EventType(eventType), payload); err != nil {
		// Increment retry count
		newRetryCount := retryCount + 1
		if newRetryCount >= MaxRetryAttempts {
			if markErr := os.markFailedTx(ctx, tx, id, fmt.Sprintf("max retries exceeded: %v", err)); markErr != nil {
				return fmt.Errorf("failed to mark event as failed: %w", markErr)
			}
			return tx.Commit()
		}

		// Update retry count
		updateQuery := fmt.Sprintf(`
			UPDATE %s 
			SET retry_count = ?, error_message = ?, last_modified_date = NOW()
			WHERE id = ?
		`, constants.TableOutboxEvent)
		if _, updateErr := tx.ExecContext(ctx, updateQuery, newRetryCount, err.Error(), id); updateErr != nil {
			return fmt.Errorf("failed to update retry count: %w", updateErr)
		}
		log.Printf("⚠️ [Outbox] Event %s failed (Attempt %d/%d). Error: %v", id, newRetryCount, MaxRetryAttempts, err)
		return tx.Commit()
	}

	// Mark as processed within the same transaction
	processedQuery := fmt.Sprintf(`
		UPDATE %s 
		SET status = ?, processed_date = NOW(), last_modified_date = NOW()
		WHERE id = ?
	`, constants.TableOutboxEvent)
	if _, err := tx.ExecContext(ctx, processedQuery, OutboxStatusProcessed, id); err != nil {
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ [Outbox] Successfully processed event %s (Type: %s)", id, eventType)
	return nil
}

// markFailedTx marks an event as failed within a transaction
func (os *OutboxService) markFailedTx(ctx context.Context, tx *sql.Tx, id, errorMsg string) error {
	query := fmt.Sprintf(`
		UPDATE %s 
		SET status = ?, error_message = ?, last_modified_date = NOW()
		WHERE id = ?
	`, constants.TableOutboxEvent)
	_, err := tx.ExecContext(ctx, query, OutboxStatusFailed, errorMsg, id)
	return err
}

// CleanupProcessed removes old processed events from the outbox.
// This should be called periodically (e.g., daily) to prevent table bloat.
func (os *OutboxService) CleanupProcessed(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)

	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE status = ? AND processed_date < ?
	`, constants.TableOutboxEvent)

	result, err := os.db.DB().ExecContext(ctx, query, OutboxStatusProcessed, cutoff)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		log.Printf("🧹 [Outbox] Cleaned up %d processed events older than %v", count, olderThan)
	}

	return count, nil
}
