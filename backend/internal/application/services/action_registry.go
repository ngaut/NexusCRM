package services

import (
	"sync"

	"github.com/nexuscrm/shared/pkg/models"
)

// ActionHandler is the interface for pluggable action handlers.
// Implement this interface to add new action types to the platform.
type ActionHandler interface {
	// Execute runs the action with the given metadata and context.
	Execute(action *models.ActionMetadata, ctx *ActionContext) error

	// Type returns the action type this handler supports.
	Type() string

	// Validate checks if the action configuration is valid.
	Validate(action *models.ActionMetadata) error
}

// ActionHandlerRegistry manages registered action handlers.
// This allows new action types to be added without modifying ActionService.
type ActionHandlerRegistry struct {
	mu       sync.RWMutex
	handlers map[string]ActionHandler
}

// NewActionHandlerRegistry creates a new empty registry
func NewActionHandlerRegistry() *ActionHandlerRegistry {
	return &ActionHandlerRegistry{
		handlers: make(map[string]ActionHandler),
	}
}

// Register adds an action handler to the registry.
// If a handler for the same type already exists, it will be replaced.
func (r *ActionHandlerRegistry) Register(handler ActionHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.Type()] = handler
}

// Get retrieves a handler for the given action type.
// Returns nil if no handler is registered.
func (r *ActionHandlerRegistry) Get(actionType string) ActionHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.handlers[actionType]
}

// Has checks if a handler is registered for the given action type.
func (r *ActionHandlerRegistry) Has(actionType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[actionType]
	return ok
}

// Types returns all registered action types.
func (r *ActionHandlerRegistry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

// DefaultRegistry is the global action handler registry.
// Use this to register custom action handlers at startup.
var DefaultRegistry = NewActionHandlerRegistry()
