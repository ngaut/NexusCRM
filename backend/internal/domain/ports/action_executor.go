package ports

import (
	"context"

	"github.com/nexuscrm/backend/internal/domain/models"
)

// ActionExecutor provides action execution capabilities.
// This interface allows testing FlowExecutor without a real ActionService.
type ActionExecutor interface {
	// ExecuteAction executes a named action by ID with the given context.
	ExecuteAction(ctx context.Context, actionID string, contextRecord models.SObject, user *models.UserSession) error

	// ExecuteActionDirect executes an action definition directly without ID lookup.
	// Used by FlowExecutor when the action configuration is already known.
	ExecuteActionDirect(ctx context.Context, action *models.ActionMetadata, record models.SObject, user *models.UserSession) error
}
