package services

import (
	"fmt"
	"time"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/formula"
)

// ServiceManager orchestrates all services with dependency injection
type ServiceManager struct {
	db *database.TiDBConnection

	// Core services
	TxManager       *TransactionManager
	EventBus        *EventBus
	Schema          *SchemaManager // Exposed for internal use
	Metadata        *MetadataService
	UIMetadata      *UIMetadataService
	Permissions     *PermissionService
	Auth            *AuthService
	QuerySvc        *QueryService
	Persistence     *PersistenceService
	ActionSvc       *ActionService
	FlowExecutor    *FlowExecutor
	FlowInstanceSvc *FlowInstanceService
	Approval        *ApprovalService
	System          *SystemManager
	Feed            *FeedService
	Notification    *NotificationService
	Validation      *ValidationService
	Outbox          *OutboxService
}

// NewServiceManager creates a new service manager with all dependencies wired
func NewServiceManager(db *database.TiDBConnection) *ServiceManager {
	sm := &ServiceManager{
		db: db,
	}

	// Initialize services in dependency order
	sm.TxManager = NewTransactionManager(db)
	sm.EventBus = NewEventBus()
	sm.Schema = NewSchemaManager(db.DB())
	sm.Metadata = NewMetadataService(db, sm.Schema)
	sm.UIMetadata = NewUIMetadataService(db, sm.Metadata)

	// Validation Service
	formulaEngine := formula.NewEngine()
	sm.Validation = NewValidationService(formulaEngine)
	sm.Metadata.SetValidationService(sm.Validation)

	sm.Permissions = NewPermissionService(db, sm.Metadata)
	sm.Metadata.SetPermissionService(sm.Permissions)
	sm.QuerySvc = NewQueryService(db, sm.Metadata, sm.Permissions)
	sm.Persistence = NewPersistenceService(db, sm.Metadata, sm.Permissions, sm.EventBus, sm.TxManager)

	// Outbox Service for transactional event storage (ACID-compliant event publishing)
	sm.Outbox = NewOutboxService(db, sm.EventBus, sm.TxManager)
	sm.Persistence.SetOutbox(sm.Outbox)

	sm.Auth = NewAuthService(db, sm.Persistence)
	sm.ActionSvc = NewActionService(sm.Metadata, sm.Persistence, sm.Permissions, sm.TxManager)

	// FlowInstanceService must be created before FlowExecutor (dependency order)
	sm.FlowInstanceSvc = NewFlowInstanceService(sm.Persistence, sm.QuerySvc, sm.Metadata)

	// FlowExecutor now takes all dependencies via constructor (no Set* methods)
	sm.FlowExecutor = NewFlowExecutor(sm.Metadata, sm.ActionSvc, sm.EventBus, sm.FlowInstanceSvc, sm.Persistence)

	sm.System = NewSystemManager(db, sm.Persistence)
	sm.Feed = NewFeedService(sm.Persistence, sm.QuerySvc)
	sm.Notification = NewNotificationService(sm.Persistence, sm.QuerySvc)

	// Approval Service
	sm.Approval = NewApprovalService(sm.Persistence, sm.QuerySvc, sm.Permissions, sm.FlowExecutor, sm.FlowInstanceSvc)

	// Register flow handlers for metadata-driven automation
	sm.FlowExecutor.RegisterFlowHandlers()

	return sm
}

// RefreshMetadataCache refreshes all metadata caches
func (sm *ServiceManager) RefreshMetadataCache() error {
	if err := sm.Metadata.RefreshCache(); err != nil {
		return fmt.Errorf("failed to refresh metadata cache: %w", err)
	}

	if err := sm.UIMetadata.RefreshCache(); err != nil {
		return fmt.Errorf("failed to refresh UI metadata cache: %w", err)
	}

	if err := sm.Permissions.RefreshPermissions(); err != nil {
		return fmt.Errorf("failed to refresh permissions: %w", err)
	}

	return nil
}

// Metadata delegation methods
func (sm *ServiceManager) GetSchema(apiName string) *models.ObjectMetadata {
	return sm.Metadata.GetSchema(apiName)
}

func (sm *ServiceManager) GetEffectiveSchema(apiName string, user *models.UserSession) *models.ObjectMetadata {
	schema := sm.Metadata.GetSchema(apiName)
	if schema == nil {
		return nil
	}
	return sm.Permissions.GetEffectiveSchema(schema, user)
}

func (sm *ServiceManager) GetSchemas() []*models.ObjectMetadata {
	return sm.Metadata.GetSchemas()
}

// App and Tab management
// Delegated directly to UIMetadata in handlers

// StartOutboxWorker starts the background outbox event processing worker.
// Call this during server startup. The worker processes pending events every 500ms.
func (sm *ServiceManager) StartOutboxWorker() {
	if sm.Outbox != nil {
		sm.Outbox.StartWorker(500 * time.Millisecond)
	}
}

// StopOutboxWorker stops the background outbox event processing worker gracefully.
// Call this during server shutdown.
func (sm *ServiceManager) StopOutboxWorker() {
	if sm.Outbox != nil {
		sm.Outbox.StopWorker()
	}
}
