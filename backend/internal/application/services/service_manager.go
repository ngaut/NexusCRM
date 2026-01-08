package services

import (
	"context"
	"fmt"
	"time"

	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/models"
)

// ServiceManager orchestrates all services with dependency injection
type ServiceManager struct {
	db *database.TiDBConnection

	// Core services
	TxManager       *persistence.TransactionManager
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
	Scheduler       *SchedulerService

	// Repositories
	UserRepo   *persistence.UserRepository
	SystemRepo *persistence.SystemRepository
}

// NewServiceManager creates a new service manager with all dependencies wired
func NewServiceManager(db *database.TiDBConnection) *ServiceManager {
	sm := &ServiceManager{
		db: db,
	}

	// 1. Infrastructure & Event Bus
	sm.TxManager = persistence.NewTransactionManager(db)
	sm.EventBus = NewEventBus()
	formulaEngine := formula.NewEngine()
	sm.Validation = NewValidationService(formulaEngine)

	// 2. Repositories (Initialize ALL Repositories first)
	schemaRepo := persistence.NewSchemaRepository(db.DB())
	sm.UserRepo = persistence.NewUserRepository(db.DB())
	sm.SystemRepo = persistence.NewSystemRepository(db.DB())
	sessionRepo := persistence.NewSessionRepository(db.DB())
	metadataRepo := persistence.NewMetadataRepository(db.DB())
	permissionRepo := persistence.NewPermissionRepository(db.DB())
	recordRepo := persistence.NewRecordRepository(db.DB())
	rollupRepo := persistence.NewRollupRepository(db.DB())
	outboxRepo := persistence.NewOutboxRepository(db.DB())
	queryRepo := persistence.NewQueryRepository(db.DB())
	schedulerRepo := persistence.NewSchedulerRepository(db.DB())

	// 3. Core Domain Managers (Foundation)
	sm.Schema = NewSchemaManager(schemaRepo)
	sm.Metadata = NewMetadataService(metadataRepo, sm.Schema)
	sm.Permissions = NewPermissionService(permissionRepo, sm.Metadata, sm.UserRepo)

	// 4. Higher-Level Orchestration Services
	sm.UIMetadata = NewUIMetadataService(sm.Metadata, sm.Permissions)
	sm.QuerySvc = NewQueryService(queryRepo, sm.Metadata, sm.Permissions)

	// 5. Persistence Ecosystem
	rollupSvc := NewRollupService(rollupRepo, sm.Metadata, sm.TxManager)
	sm.Outbox = NewOutboxService(outboxRepo, sm.EventBus, sm.TxManager)

	sm.Persistence = NewPersistenceService(
		recordRepo,
		rollupSvc,
		sm.Metadata,
		sm.Permissions,
		sm.EventBus,
		sm.Validation,
		sm.TxManager,
		sm.Outbox,
	)

	// 6. Business Logic Services
	sm.ActionSvc = NewActionService(sm.Metadata, sm.Persistence, sm.Permissions, sm.TxManager)

	// Flow Stack (Order matters: Instance -> Executor)
	sm.FlowInstanceSvc = NewFlowInstanceService(sm.Persistence, sm.QuerySvc, sm.Metadata)
	sm.FlowExecutor = NewFlowExecutor(sm.Metadata, sm.ActionSvc, sm.EventBus, sm.FlowInstanceSvc, sm.Persistence)
	// Register flow handlers
	sm.FlowExecutor.RegisterFlowHandlers()

	sm.System = NewSystemManager(sm.Persistence, sm.SystemRepo)
	sm.Feed = NewFeedService(sm.Persistence, sm.QuerySvc)
	sm.Notification = NewNotificationService(sm.Persistence, sm.QuerySvc)

	// Approval Service
	sm.Approval = NewApprovalService(sm.Persistence, sm.QuerySvc, sm.Permissions, sm.FlowExecutor, sm.FlowInstanceSvc)

	// Scheduler Service
	sm.Scheduler = NewSchedulerService(schedulerRepo, sm.Metadata, sm.FlowExecutor)

	// 7. Auth Service (Instantiated last to satisfy dependencies)
	sm.Auth = NewAuthService(sm.Persistence, sm.UserRepo, sessionRepo, permissionRepo)

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
func (sm *ServiceManager) GetSchema(ctx context.Context, apiName string) *models.ObjectMetadata {
	return sm.Metadata.GetSchema(ctx, apiName)
}

func (sm *ServiceManager) GetEffectiveSchema(ctx context.Context, apiName string, user *models.UserSession) *models.ObjectMetadata {
	schema := sm.Metadata.GetSchema(ctx, apiName)
	if schema == nil {
		return nil
	}
	return sm.Permissions.GetEffectiveSchema(ctx, schema, user)
}

func (sm *ServiceManager) GetSchemas(ctx context.Context) []*models.ObjectMetadata {
	return sm.Metadata.GetSchemas(ctx)
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

// StartScheduler starts the scheduled job executor.
// Call this during server startup.
func (sm *ServiceManager) StartScheduler() {
	if sm.Scheduler != nil {
		go sm.Scheduler.Start()
	}
}

// StopScheduler stops the scheduled job executor gracefully.
// Call this during server shutdown.
func (sm *ServiceManager) StopScheduler() {
	if sm.Scheduler != nil {
		sm.Scheduler.Stop()
	}
}
