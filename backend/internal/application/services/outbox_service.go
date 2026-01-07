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
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
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
	repo      *persistence.OutboxRepository
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
		repo:      persistence.NewOutboxRepository(db.DB()),
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
	id, err := os.repo.Enqueue(ctx, tx, string(eventType), payload)
	if err != nil {
		return err
	}
	log.Printf("✅ [Outbox] Enqueued event %s (Type: %s, ID: %s)", eventType, string(eventType), id)
	return nil
}

// enqueueDirect inserts event directly without transaction context
func (os *OutboxService) enqueueDirect(ctx context.Context, eventType events.EventType, payload RecordEventPayload) error {
	id, err := os.repo.Enqueue(ctx, os.db.DB(), string(eventType), payload)
	if err != nil {
		return err
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
	// First, get pending events
	events, err := os.repo.GetPendingEvents(ctx, 100)
	if err != nil {
		return err
	}

	if len(events) > 0 {
		log.Printf("🔄 [Outbox] Processing %d pending events", len(events))
	}

	// Process each event atomically (claim, process, update status)
	for _, e := range events {
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
	// Try to claim this specific event
	claimedID, err := os.repo.ClaimEvent(ctx, tx, id)
	if err != nil {
		return fmt.Errorf("failed to claim event: %w", err)
	}
	if claimedID == "" {
		return nil // Already processed/locked
	}

	// Parse payload
	// Parse payload
	var payload RecordEventPayload
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		log.Printf("❌ [Outbox] Event %s failed payload unmarshal: %v", id, err)
		if markErr := os.repo.UpdateStatus(ctx, tx, id, OutboxStatusFailed, fmt.Sprintf("invalid payload: %v", err)); markErr != nil {
			return fmt.Errorf("failed to mark event as failed: %w", markErr)
		}
		return tx.Commit()
	}

	// Publish via EventBus
	// Publish via EventBus
	if err := os.eventBus.Publish(ctx, events.EventType(eventType), payload); err != nil {
		// Increment retry count
		newRetryCount := retryCount + 1
		if newRetryCount >= MaxRetryAttempts {
			if markErr := os.repo.UpdateStatus(ctx, tx, id, OutboxStatusFailed, fmt.Sprintf("max retries exceeded: %v", err)); markErr != nil {
				return fmt.Errorf("failed to mark event as failed: %w", markErr)
			}
			return tx.Commit()
		}

		// Update retry count
		if updateErr := os.repo.IncrementRetry(ctx, tx, id, newRetryCount, err.Error()); updateErr != nil {
			return fmt.Errorf("failed to update retry count: %w", updateErr)
		}
		log.Printf("⚠️ [Outbox] Event %s failed (Attempt %d/%d). Error: %v", id, newRetryCount, MaxRetryAttempts, err)
		return tx.Commit()
	}

	// Mark as processed within the same transaction
	// Mark as processed within the same transaction
	if err := os.repo.UpdateStatus(ctx, tx, id, OutboxStatusProcessed, ""); err != nil {
		return fmt.Errorf("failed to mark as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("✅ [Outbox] Successfully processed event %s (Type: %s)", id, eventType)
	return nil
}

// CleanupProcessed removes old processed events from the outbox.
// This should be called periodically (e.g., daily) to prevent table bloat.
func (os *OutboxService) CleanupProcessed(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	return os.repo.CleanupProcessed(ctx, cutoff)
}
