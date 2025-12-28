package ports

import (
	"context"

	"github.com/nexuscrm/backend/internal/domain/events"
)

// EventHandler is a function that handles an event
type EventHandler func(ctx context.Context, payload interface{}) error

// EventPublisher provides event publishing capabilities.
// Implementations should handle async event dispatching.
type EventPublisher interface {
	// Subscribe registers a handler for a specific event type.
	Subscribe(eventType events.EventType, handler EventHandler) func()

	// Publish dispatches an event to all registered handlers.
	// Returns an error if any handler fails.
	Publish(ctx context.Context, eventType events.EventType, payload interface{}) error
}
