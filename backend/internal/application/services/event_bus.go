package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/shared/pkg/models"
	"github.com/nexuscrm/backend/internal/domain/ports"
)

// EventType is an alias to the domain type
type EventType = events.EventType

// RecordEventPayload represents payload for record events
type RecordEventPayload struct {
	ObjectAPIName string              `json:"object_api_name"`
	Record        models.SObject      `json:"record"`
	OldRecord     *models.SObject     `json:"old_record,omitempty"`
	CurrentUser   *models.UserSession `json:"current_user,omitempty"`
}

// PlatformEvent represents a platform event
type PlatformEvent struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}

// EventHandler is a function that handles an event.
// Using the type from ports to ensure interface compatibility.
type EventHandler = ports.EventHandler

// EventBus manages publish-subscribe event system.
// It implements ports.EventPublisher interface.
type EventBus struct {
	handlers map[EventType][]EventHandler
	mu       sync.RWMutex
}

// Ensure EventBus implements ports.EventPublisher at compile time
var _ ports.EventPublisher = (*EventBus)(nil)

// NewEventBus creates a new EventBus instance
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]EventHandler),
	}
}

// Subscribe registers a handler for a specific event type
// Returns an unsubscribe function
func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.handlers[eventType] == nil {
		eb.handlers[eventType] = make([]EventHandler, 0)
	}

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)

	// Return unsubscribe function
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()

		handlers := eb.handlers[eventType]
		for i, h := range handlers {
			// Compare function pointers
			if &h == &handler {
				eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// Publish publishes an event to all registered handlers
func (eb *EventBus) Publish(ctx context.Context, eventType EventType, payload interface{}) error {
	eb.mu.RLock()
	handlers := eb.handlers[eventType]
	eb.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	// Create platform event
	event := PlatformEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().Unix(),
	}

	// Execute handlers in sequence
	for _, handler := range handlers {
		if err := handler(ctx, event.Payload); err != nil {
			return fmt.Errorf("EventBus handler error for %s: %w", eventType, err)
		}
	}

	return nil
}

// PublishAsync publishes an event asynchronously
func (eb *EventBus) PublishAsync(eventType EventType, payload interface{}) {
	go func() {
		// Use background context for async events as they are decoupled from the request/tx
		if err := eb.Publish(context.Background(), eventType, payload); err != nil {
			log.Printf("EventBus async publish error: %v", err)
		}
	}()
}

// Clear removes all handlers (useful for testing)
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers = make(map[EventType][]EventHandler)
}
