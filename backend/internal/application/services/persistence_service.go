package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/infrastructure/database"
	"github.com/nexuscrm/backend/pkg/constants"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/backend/pkg/query"
)

// PersistenceService handles CRUD operations with validation and events
type PersistenceService struct {
	db          *database.TiDBConnection
	metadata    *MetadataService
	permissions *PermissionService
	eventBus    *EventBus
	formula     *formula.Engine
	validator   *ValidationService
	txManager   *TransactionManager
	rollup      *RollupService
	outbox      *OutboxService
}

// NewPersistenceService creates a new PersistenceService
func NewPersistenceService(
	db *database.TiDBConnection,
	metadata *MetadataService,
	permissions *PermissionService,
	eventBus *EventBus,
	txManager *TransactionManager,
) *PersistenceService {
	// Initialize RollupService
	rollup := NewRollupService(db.DB(), metadata, txManager)

	return &PersistenceService{
		db:          db,
		metadata:    metadata,
		permissions: permissions,
		eventBus:    eventBus,
		formula:     formula.NewEngine(),
		validator:   NewValidationService(formula.NewEngine()),
		txManager:   txManager,
		rollup:      rollup,
	}
}

// SetOutbox sets the OutboxService for transactional event storage.
// This is called after construction to avoid circular dependencies.
func (ps *PersistenceService) SetOutbox(outbox *OutboxService) {
	ps.outbox = outbox
}

// ==================== CRUD Operations ====================

// publishRecordEvent publishes a record event with consistent payload
func (ps *PersistenceService) publishRecordEvent(ctx context.Context, eventType events.EventType, objectName string, record models.SObject, oldRecord *models.SObject, currentUser *models.UserSession) error {
	payload := RecordEventPayload{
		ObjectAPIName: objectName,
		Record:        record,
		OldRecord:     oldRecord,
		CurrentUser:   currentUser,
	}

	if err := ps.eventBus.Publish(ctx, eventType, payload); err != nil {
		return fmt.Errorf("%s event failed: %w", eventType, err)
	}
	return nil
}

// prepareOperation checks permissions and retrieves schema
func (ps *PersistenceService) prepareOperation(objectName string, operation string, user *models.UserSession) (*models.ObjectMetadata, error) {
	if err := ps.permissions.CheckPermissionOrErrorWithUser(objectName, operation, user); err != nil {
		return nil, err
	}
	return ps.metadata.GetSchemaOrError(objectName)
}

// validatePolymorphicLookups verifies that referenced IDs in polymorphic fields exist in at least one of the allowed objects
// Returns a map of fieldName -> foundObjectType
func (ps *PersistenceService) validatePolymorphicLookups(ctx context.Context, data models.SObject, schema *models.ObjectMetadata) (map[string]string, error) {
	resolved := make(map[string]string)
	for _, field := range schema.Fields {
		// Only check polymorphic lookups that have a value
		if field.Type != constants.FieldTypeLookup || !field.IsPolymorphic || len(field.ReferenceTo) == 0 {
			continue
		}

		val, ok := data[field.APIName]
		if !ok || val == nil {
			continue
		}

		idVal, ok := val.(string)
		if !ok || idVal == "" {
			continue
		}

		// Check if ID exists in ANY of the referenced objects
		found := false
		var checkedObjects []string

		for _, refObj := range field.ReferenceTo {
			checkedObjects = append(checkedObjects, refObj)
			// Efficient existence check using ID
			exists, err := ps.checkRecordExists(ctx, refObj, idVal)
			if err != nil {
				log.Printf("Warning: failed to check existence in %s: %v", refObj, err)
				continue
			}
			if exists {
				found = true
				resolved[field.APIName] = refObj
				break
			}
		}

		if !found {
			return nil, errors.NewValidationError(field.APIName, fmt.Sprintf("referenced ID %s not found in any of the allowed objects: %s", idVal, strings.Join(checkedObjects, ", ")))
		}
	}
	return resolved, nil
}

// checkRecordExists checks if a record exists by ID (bypassing permissions for validation)
func (ps *PersistenceService) checkRecordExists(ctx context.Context, objectName string, id string) (bool, error) {
	queryP := query.From(objectName).
		Select([]string{constants.FieldID}).
		Where(constants.FieldID+" = ?", id).
		Limit(1).
		Build()

	rows, err := ps.db.DB().QueryContext(ctx, queryP.SQL, queryP.Params...)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

// RunInTransaction executes a function within a transaction (new or existing) with retry support
func (ps *PersistenceService) RunInTransaction(
	ctx context.Context,
	fn func(tx *sql.Tx, ctx context.Context) error,
) error {
	existingTx := ps.txManager.ExtractTx(ctx)

	// If already in a transaction, execute directly
	if existingTx != nil {
		return fn(existingTx, ctx)
	}

	// Otherwise start new transaction with retry
	return ps.txManager.WithRetry(func(tx *sql.Tx) error {
		// Inject transaction into context for downstream usage
		txCtx := ps.txManager.InjectTx(ctx, tx)
		return fn(tx, txCtx)
	}, 3)
}

// Insert operations are in persistence_insert.go:
// - Insert, processMentions

// Update modifies an existing record with ACID transaction guarantees
func (ps *PersistenceService) Update(
	ctx context.Context,
	objectName string,
	id string,
	updates models.SObject,
	currentUser *models.UserSession,
) error {
	schema, err := ps.prepareOperation(objectName, constants.PermEdit, currentUser)
	if err != nil {
		return err
	}

	var finalRecord models.SObject
	var oldRecord models.SObject

	// Execute Transactional Work
	err = ps.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Load current record with FOR UPDATE lock within transaction
		q := query.From(objectName).
			Select([]string{"*"}).
			Where(fmt.Sprintf("%s = ?", constants.FieldID), id).
			Limit(1).
			Build()

		// Inject FOR UPDATE for locking
		q.SQL += " FOR UPDATE"

		oldRecords, err := ExecuteQuery(txCtx, tx, q)
		if err != nil {
			return fmt.Errorf("failed to lock record: %w", err)
		}

		if len(oldRecords) == 0 {
			return errors.NewNotFoundError(objectName, id)
		}

		oldRecord = oldRecords[0]

		// Validate Polymorphic Lookups & Resolve Types
		resolvedTypes, err := ps.validatePolymorphicLookups(txCtx, updates, schema)
		if err != nil {
			return err
		}

		// Inject resolved types into updates for persistence
		for fieldName, objType := range resolvedTypes {
			updates[GetPolymorphicTypeColumnName(fieldName)] = objType
		}

		// Check record-level access
		if !ps.permissions.CheckRecordAccess(schema, oldRecord, constants.PermEdit, currentUser) {
			return errors.NewPermissionError("update", objectName+"/"+id)
		}

		// Filter editable fields
		effectiveUpdates := make(models.SObject)
		hasChanges := false

		// Normalize updates (keys to match schema, values to match type)
		normalizedUpdates := NormalizeSObject(schema, updates)
		updates = normalizedUpdates

		for key, newVal := range updates {
			if isFieldSystemReadOnly(ps.metadata, objectName, key) {
				continue
			}

			if !ps.permissions.CheckFieldEditabilityWithUser(objectName, key, currentUser) {
				continue
			}

			oldVal := oldRecord[key]
			if !ps.areValuesEqual(oldVal, newVal) {
				effectiveUpdates[key] = newVal
				hasChanges = true
			}
		}

		if !hasChanges {
			return nil // No changes
		}

		// Merge for validation
		recordToValidate := ps.mergeRecords(oldRecord, effectiveUpdates)

		// Validate
		validationRules := ps.metadata.GetValidationRules(objectName)
		if err := ps.validator.ValidateRecord(recordToValidate, schema, validationRules, &oldRecord); err != nil {
			return err
		}

		// Check uniqueness
		if err := ps.checkUniqueness(txCtx, objectName, effectiveUpdates, schema, id); err != nil {
			return err
		}

		// Publish beforeUpdate event
		if err := ps.publishRecordEvent(txCtx, events.RecordBeforeUpdate, objectName, recordToValidate, &oldRecord, currentUser); err != nil {
			return err
		}

		// Re-calculate effectiveUpdates to capture any changes made by Flows/Triggers
		for key, newVal := range recordToValidate {
			oldVal := oldRecord[key]
			if !ps.areValuesEqual(oldVal, newVal) {
				effectiveUpdates[key] = newVal
			}
		}

		// Add system fields for update (lastModifiedDate, lastModifiedById)
		systemFieldsUpdate := ps.generateSystemFields(objectName, effectiveUpdates, currentUser, false)
		for k, v := range systemFieldsUpdate {
			effectiveUpdates[k] = v
		}

		// Generate Audit Logs (excluding system objects to prevent recursion)
		if !constants.IsSystemTable(objectName) || objectName == constants.TableUser {
			for key, newVal := range effectiveUpdates {
				// Skip system fields (audit log should track user changes)
				if isFieldSystemReadOnly(ps.metadata, objectName, key) {
					continue
				}

				// Create Audit Log Entry
				auditID := fmt.Sprintf("%s-%d-%s", id, time.Now().UnixNano(), key)
				oldVal := oldRecord[key]

				auditEntryStruct := models.SystemAuditLog{
					ID:            auditID,
					ObjectAPIName: objectName,
					RecordID:      id,
					FieldName:     key,
					OldValue:      ps.valToString(oldVal),
					NewValue:      ps.valToString(newVal),
					ChangedByID:   currentUser.ID,
					ChangedAt:     time.Now(),
				}
				auditEntry := auditEntryStruct.ToSObject()

				// Use raw query builder for system table insert to avoid recursion/permission checks
				// We manually build insert for _System_AuditLog
				auditQ := query.Insert(constants.TableAuditLog, auditEntry).Build()
				if _, err := tx.Exec(auditQ.SQL, auditQ.Params...); err != nil {
					// We log but don't fail transaction for audit failure?
					// Ideally we SHOULD fail for strict audit.
					return fmt.Errorf("failed to write audit log: %w", err)
				}
			}
		}

		// Extract physical fields
		physicalUpdates := ToStorageRecord(schema, effectiveUpdates)

		// Build and execute update within transaction
		builder := query.Update(objectName).
			Set(physicalUpdates).
			Where(fmt.Sprintf("%s = ?", constants.FieldID), id)

		q = builder.Build()

		if _, err := tx.Exec(q.SQL, q.Params...); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		// Set final record for afterUpdate event
		finalRecord = ps.mergeRecords(recordToValidate, effectiveUpdates)

		// Hook: Rollup Summary (inside transaction for ACID compliance)
		// Process rollups for NEW record state
		if err := ps.rollup.ProcessRollups(txCtx, tx, objectName, finalRecord); err != nil {
			return fmt.Errorf("failed to process rollups: %w", err)
		}

		// Also process rollups for OLD record state (handles parent changes)
		if err := ps.rollup.ProcessRollups(txCtx, tx, objectName, oldRecord); err != nil {
			return fmt.Errorf("failed to process old record rollups: %w", err)
		}

		// Enqueue afterUpdate event to outbox (inside transaction for guaranteed delivery)
		if ps.outbox != nil {
			if err := ps.outbox.EnqueueEventTx(txCtx, tx, events.RecordUpdated, RecordEventPayload{
				ObjectAPIName: objectName,
				Record:        finalRecord,
				OldRecord:     &oldRecord,
				CurrentUser:   currentUser,
			}); err != nil {
				return fmt.Errorf("failed to enqueue record updated event: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	updatedID, _ := finalRecord[constants.FieldID].(string)
	log.Printf("📝 Updated record %s in %s (User: %s)", updatedID, objectName, getUserID(currentUser))

	return nil
}

// valToString converts any value to string for audit logging
func (ps *PersistenceService) valToString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case *time.Time:
		if v != nil {
			return v.Format(time.RFC3339)
		}
		return ""
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// Delete and cascadeDeleteChildren are in persistence_delete.go

// Helper to safely get user ID
func getUserID(user *models.UserSession) string {
	if user == nil {
		return "system"
	}
	return user.ID
}
