package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nexuscrm/backend/internal/domain/events"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/backend/pkg/formula"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// PersistenceService manages strict CRUD operations and data integrity
type PersistenceService struct {
	repo        *persistence.RecordRepository
	metadata    *MetadataService
	permissions *PermissionService
	eventBus    *EventBus
	formula     *formula.Engine
	validator   *ValidationService
	txManager   *persistence.TransactionManager
	rollup      *RollupService
	outbox      *OutboxService
}

// NewPersistenceService creates a new PersistenceService
func NewPersistenceService(
	repo *persistence.RecordRepository,
	rollup *RollupService,
	metadata *MetadataService,
	permissions *PermissionService,
	eventBus *EventBus,
	validator *ValidationService,
	txManager *persistence.TransactionManager,
	outbox *OutboxService,
) *PersistenceService {
	return &PersistenceService{
		repo:        repo,
		rollup:      rollup,
		metadata:    metadata,
		permissions: permissions,
		eventBus:    eventBus,
		validator:   validator,
		txManager:   txManager,
		formula:     formula.NewEngine(),
		outbox:      outbox,
	}
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
func (ps *PersistenceService) prepareOperation(ctx context.Context, objectName string, operation string, user *models.UserSession) (*models.ObjectMetadata, error) {
	if err := ps.permissions.CheckPermissionOrErrorWithUser(ctx, objectName, operation, user); err != nil {
		return nil, err
	}
	return ps.metadata.GetSchemaOrError(ctx, objectName)
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
			tableToCheck := ps.getTableName(refObj) // Resolving table name
			checkedObjects = append(checkedObjects, tableToCheck)
			// Efficient existence check using ID
			exists, err := ps.checkRecordExists(ctx, tableToCheck, idVal)
			log.Printf("🔍 Poly check: Field=%s, ID=%s, Obj=%s, Table=%s, Exists=%v, Err=%v", field.APIName, idVal, refObj, tableToCheck, exists, err)
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
	// Attempt to extract transaction, though usually this is called within one or it handles nil fine
	tx := ps.txManager.ExtractTx(ctx)
	// IMPORTANT: table name resolution happens here or inside repo?
	// Given we are moving to infrastructure, using objectName as table name is standard for now unless we have mapping.
	// PersistenceService.getTableName() is still useful.
	tableName := ps.getTableName(objectName)

	return ps.repo.Exists(ctx, tx, tableName, id)
}

// Helper to get table name from object API name
// Helper to get table name from object API name
func (ps *PersistenceService) getTableName(objectAPIName string) string {
	// Standard mapping: APIName matches TableName
	return objectAPIName
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
	schema, err := ps.prepareOperation(ctx, objectName, constants.PermEdit, currentUser)
	if err != nil {
		return err
	}

	var finalRecord models.SObject
	var oldRecord models.SObject

	// Execute Transactional Work
	err = ps.RunInTransaction(ctx, func(tx *sql.Tx, txCtx context.Context) error {
		// Load current record with FOR UPDATE lock within transaction via Repository
		oldRecord, err = ps.repo.GetLock(txCtx, tx, objectName, id)
		if err != nil {
			return fmt.Errorf("failed to lock record: %w", err)
		}
		if oldRecord == nil {
			return errors.NewNotFoundError(objectName, id)
		}

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
		if !ps.permissions.CheckRecordAccess(ctx, schema, oldRecord, constants.PermEdit, currentUser) {
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

			if !ps.permissions.CheckFieldEditabilityWithUser(ctx, objectName, key, currentUser) {
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
		validationRules := ps.metadata.GetValidationRules(txCtx, objectName)
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
		systemFieldsUpdate := ps.generateSystemFields(ctx, objectName, effectiveUpdates, currentUser, false)
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
					ID:               auditID,
					ObjectAPIName:    objectName,
					RecordID:         id,
					FieldName:        key,
					OldValue:         ps.valToString(oldVal),
					NewValue:         ps.valToString(newVal),
					ChangedByID:      currentUser.ID,
					ChangedAt:        time.Now(),
					CreatedDate:      time.Now(),
					LastModifiedDate: time.Now(),
				}
				auditEntry := auditEntryStruct.ToSObject()

				// Use Repo for System Audit Insert
				// Note: Audit Log table is dynamic/system, we can use same Insert pattern
				if err := ps.repo.Insert(txCtx, tx, constants.TableAuditLog, auditEntry); err != nil {
					return fmt.Errorf("failed to write audit log: %w", err)
				}
			}
		}

		// Extract physical fields
		physicalUpdates := ToStorageRecord(schema, effectiveUpdates)

		// execute update via Repository
		if err := ps.repo.Update(txCtx, tx, objectName, id, physicalUpdates); err != nil {
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

// generateAutoNumbers handles atomic generation of AutoNumber field values within a transaction
func (ps *PersistenceService) generateAutoNumbers(ctx context.Context, tx *sql.Tx, objectName string, data models.SObject) error {
	autoNumbers := ps.metadata.GetAutoNumbers(ctx, objectName)
	if len(autoNumbers) == 0 {
		return nil
	}

	for _, an := range autoNumbers {
		// 1. Lock and increment current value in DB via Repo
		// Can't use GetLock for standard SObject because this is _System_AutoNumber
		// But RecordRepository is generic, so we can use it!
		// GetLock returns SObject.

		autoNumberRecord, err := ps.repo.GetLock(ctx, tx, constants.TableAutoNumber, an.ID)
		if err != nil {
			return fmt.Errorf("failed to lock auto-number %s: %w", an.ID, err)
		}

		var currentValue int
		if autoNumberRecord != nil {
			if val, ok := autoNumberRecord["current_number"]; ok && val != nil {
				// Handle potential float64 from JSON/DB driver
				switch v := val.(type) {
				case int64:
					currentValue = int(v)
				case float64:
					currentValue = int(v)
				case int:
					currentValue = v
				}
			}
		}

		newValue := currentValue + 1

		// 2. Update DB via Repo
		anUpdate := models.SObject{
			constants.FieldSysAutoNumber_CurrentNumber: newValue,
			constants.FieldLastModifiedDate:            time.Now().UTC(),
		}

		if err := ps.repo.Update(ctx, tx, constants.TableAutoNumber, an.ID, anUpdate); err != nil {
			return fmt.Errorf("failed to update auto-number %s: %w", an.ID, err)
		}

		// 3. Format value
		formatted := ps.formatAutoNumber(an.DisplayFormat, newValue)

		// 4. Set value in data
		data[an.FieldAPIName] = formatted
	}
	return nil
}

// formatAutoNumber applies a display format (e.g., "INV-{0000}") to a numeric value
func (ps *PersistenceService) formatAutoNumber(format string, value int) string {
	start := strings.Index(format, "{")
	end := strings.Index(format, "}")
	if start == -1 || end == -1 || end <= start {
		// Fallback: just append
		return fmt.Sprintf("%s%d", format, value)
	}

	placeholder := format[start+1 : end]
	padding := len(placeholder)

	formattedValue := fmt.Sprintf("%0*d", padding, value)
	return format[:start] + formattedValue + format[end+1:]
}
