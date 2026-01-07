package ports

import (
	"context"

	"github.com/nexuscrm/shared/pkg/models"
)

// FlowMetadataProvider provides access to flow definitions.
// This interface enables testing FlowExecutor without a real MetadataService.
type FlowMetadataProvider interface {
	// GetFlows returns all flow definitions.
	GetFlows(ctx context.Context) []*models.Flow

	// GetSupportedEvents returns the list of event types that flows can trigger on.
	GetSupportedEvents() []string

	// GetFlow returns a specific flow by ID.
	GetFlow(ctx context.Context, id string) *models.Flow
}

// MetadataProvider provides comprehensive metadata access.
// This is a superset of FlowMetadataProvider for services that need full metadata.
type MetadataProvider interface {
	FlowMetadataProvider

	// GetSchema returns the metadata for a specific object.
	GetSchema(ctx context.Context, apiName string) *models.ObjectMetadata

	// GetSchemas returns all object metadata.
	GetSchemas(ctx context.Context) []*models.ObjectMetadata

	// GetAction retrieves a specific action by ID.
	GetAction(ctx context.Context, actionID string) *models.ActionMetadata
}

// FlowStepExecutor executes a specific step within a flow instance.
// This interface enables FlowInstanceService to trigger step execution without circular dependency.
type FlowStepExecutor interface {
	ExecuteInstanceStep(ctx context.Context, instance *models.FlowInstance, step *models.FlowStep, record models.SObject, user *models.UserSession) error
}
