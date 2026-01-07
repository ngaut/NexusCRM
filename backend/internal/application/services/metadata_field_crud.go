package services

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	domainSchema "github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Field CRUD Methods ====================

// CreateField creates a new field for an object
func (ms *MetadataService) CreateField(objectAPIName string, field *models.FieldMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate Field Metadata
	if ms.validationSvc != nil {
		if err := ms.validationSvc.ValidateFieldMetadata(field); err != nil {
			return err
		}
	}

	// Validate API Name format
	// Allow alphanumeric and underscores, must start with letter/underscore
	validName := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !validName.MatchString(field.APIName) {
		return fmt.Errorf("invalid field API name '%s': must start with letter or underscore and contain only alphanumeric characters", field.APIName)
	}

	// Master-Detail Validation/Enforcement
	if field.IsMasterDetail {
		if field.Type != constants.FieldTypeLookup {
			return errors.NewValidationError(constants.FieldMetaType, "Master-Detail relationship is only valid for Lookup fields")
		}
		// Enforce defaults for Master-Detail
		field.Required = true
		cascade := constants.DeleteRuleCascade
		field.DeleteRule = &cascade
	}

	// Validate System Field restrictions
	if constants.IsSystemField(field.APIName) {
		return fmt.Errorf("cannot create field with reserved system name '%s'", field.APIName)
	}

	// Detect polymorphic lookups early
	if field.Type == constants.FieldTypeLookup && len(field.ReferenceTo) > 1 {
		field.IsPolymorphic = true
	}

	// Validate Picklist fields require options
	if field.Type == constants.FieldTypePicklist && len(field.Options) == 0 {
		return errors.NewValidationError("options", "Picklist fields require at least one option")
	}

	// Validate Lookup fields require reference_to
	if field.Type == constants.FieldTypeLookup && len(field.ReferenceTo) == 0 {
		return errors.NewValidationError("reference_to", "Lookup fields require a referenced object")
	}

	// Get the object to ensure it exists
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), objectAPIName)
	if err != nil || obj == nil {
		return fmt.Errorf("object '%s' not found", objectAPIName)
	}

	// Validate Max Master-Detail Usage (Limit 2)
	if field.IsMasterDetail {
		if err := ms.ValidateMaxMasterDetailFields(obj, ""); err != nil {
			return err
		}
	}

	// Check if field already exists
	for _, f := range obj.Fields {
		if strings.EqualFold(f.APIName, field.APIName) {
			return fmt.Errorf("field '%s' already exists on object '%s'", field.APIName, objectAPIName)
		}
	}

	// Validate Formula fields
	if field.Type == constants.FieldTypeFormula {
		if field.Formula == nil || *field.Formula == "" {
			return errors.NewValidationError("formula", "Formula fields require a formula expression")
		}
		if field.ReturnType == nil || string(*field.ReturnType) == "" {
			return errors.NewValidationError("return_type", "Formula fields require a valid return_type")
		}
		// Validate formula syntax by attempting to compile it
		sampleEnv := make(map[string]interface{})
		for _, f := range obj.Fields {
			switch f.Type {
			case constants.FieldTypeNumber, constants.FieldTypeCurrency, constants.FieldTypePercent:
				sampleEnv[f.APIName] = 0.0
			case constants.FieldTypeBoolean:
				sampleEnv[f.APIName] = false
			default:
				sampleEnv[f.APIName] = ""
			}
		}
		if err := ms.schemaMgr.ValidateFormula(*field.Formula, sampleEnv); err != nil {
			return errors.NewValidationError("formula", fmt.Sprintf("Invalid formula syntax: %v", err))
		}
	}

	// Map to ColumnDefinition
	var relationshipName string
	if field.RelationshipName != nil {
		relationshipName = *field.RelationshipName
	}

	colDef := domainSchema.ColumnDefinition{
		Name:             field.APIName,
		Type:             ms.schemaMgr.MapFieldTypeToSQL(string(field.Type)),
		LogicalType:      string(field.Type),
		Nullable:         !field.Required,
		Unique:           field.Unique,
		IsMasterDetail:   field.IsMasterDetail,
		RelationshipName: relationshipName,
	}
	if field.DefaultValue != nil {
		colDef.Default = "'" + *field.DefaultValue + "'"
	}
	if len(field.ReferenceTo) == 1 {
		colDef.ReferenceTo = field.ReferenceTo[0]
	}
	colDef.AllReferences = field.ReferenceTo
	if field.Formula != nil {
		colDef.Formula = *field.Formula
	}
	if field.ReturnType != nil {
		colDef.ReturnType = string(*field.ReturnType)
	}
	if field.DeleteRule != nil {
		if *field.DeleteRule == constants.DeleteRuleCascade {
			colDef.OnDelete = "CASCADE"
		} else if *field.DeleteRule == constants.DeleteRuleSetNull {
			colDef.OnDelete = "SET NULL"
		} else {
			colDef.OnDelete = "RESTRICT"
		}
	}
	if len(field.Options) > 0 {
		colDef.Options = field.Options
	}

	// Delegate to SchemaManager
	if err := ms.schemaMgr.AddColumn(objectAPIName, colDef); err != nil {
		return fmt.Errorf("failed to add column to schema: %w", err)
	}

	// For Polymorphic Lookups, create a secondary column for Object Type
	if field.IsPolymorphic {
		typeColName := GetPolymorphicTypeColumnName(field.APIName)
		typeColDef := domainSchema.ColumnDefinition{
			Name:        typeColName,
			Type:        "VARCHAR(100)",
			LogicalType: string(constants.FieldTypeText),
			Nullable:    true,
		}
		if err := ms.schemaMgr.AddColumn(objectAPIName, typeColDef); err != nil {
			log.Printf("🔥 Failed to add polymorphic type column %s. Rolling back primary column...", typeColName)
			if dropErr := ms.schemaMgr.DropColumn(objectAPIName, field.APIName); dropErr != nil {
				log.Printf("⚠️ Rollback failed for column %s: %v", field.APIName, dropErr)
			}
			return fmt.Errorf("failed to add type column for polymorphic field: %w", err)
		}
	}

	// Add to default layout
	if err := ms.addFieldToLayout(objectAPIName, field.APIName); err != nil {
		log.Printf("Warning: Failed to add field %s to layout: %v", field.APIName, err)
	}

	// AUTOMATIC PERMISSION GRANT
	if ms.permissionSvc != nil {
		profiles := []string{constants.ProfileSystemAdmin, constants.ProfileStandardUser}
		for _, profileID := range profiles {
			fieldPerm := models.FieldPermission{
				ProfileID:     &profileID,
				ObjectAPIName: objectAPIName,
				FieldAPIName:  field.APIName,
				Readable:      true,
				Editable:      true,
			}
			if err := ms.permissionSvc.UpdateFieldPermission(fieldPerm); err != nil {
				log.Printf("⚠️ Failed to grant default permission for field %s to profile %s: %v", field.APIName, profileID, err)
			}

			if field.IsPolymorphic {
				typeFieldPerm := fieldPerm
				typeFieldPerm.FieldAPIName = GetPolymorphicTypeColumnName(field.APIName)
				if err := ms.permissionSvc.UpdateFieldPermission(typeFieldPerm); err != nil {
					log.Printf("⚠️ Failed to grant permission for polymorphic type field: %v", err)
				}
			}
		}
	}

	// For AutoNumber fields, register in _System_AutoNumber
	if field.Type == constants.FieldTypeAutoNumber {
		format := "{0}"
		if field.DefaultValue != nil && *field.DefaultValue != "" {
			format = *field.DefaultValue
		}

		anID := GenerateAutoNumberID(objectAPIName, field.APIName)
		// Default starting_number to 1 (current_number = 0)
		if err := ms.repo.UpsertAutoNumber(context.Background(), anID, objectAPIName, field.APIName, format, 1, 0); err != nil {
			log.Printf("⚠️ Failed to register auto-number metadata: %v", err)
		}
	}

	ms.invalidateCacheLocked()
	return nil
}

// UpdateField updates an existing field
func (ms *MetadataService) UpdateField(objectAPIName, fieldAPIName string, updates *models.FieldMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get the object
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), objectAPIName)
	if err != nil || obj == nil {
		return fmt.Errorf("object '%s' not found", objectAPIName)
	}

	// Find the field
	var existingField *models.FieldMetadata
	for i := range obj.Fields {
		if obj.Fields[i].APIName == fieldAPIName {
			existingField = &obj.Fields[i]
			break
		}
	}
	if existingField == nil {
		return fmt.Errorf("field '%s' not found on object '%s'", fieldAPIName, objectAPIName)
	}

	// Don't allow editing system fields (except help text and default value which is safe)
	if existingField.IsSystem {
		if (updates.Type != "" && updates.Type != existingField.Type) ||
			(updates.Required != existingField.Required) ||
			(updates.Unique != existingField.Unique) {
			return fmt.Errorf("cannot modify structural properties of system field '%s'", fieldAPIName)
		}
	}

	// Apply updates
	if updates.Label != "" {
		existingField.Label = updates.Label
	}
	existingField.Required = updates.Required
	existingField.Unique = updates.Unique

	if updates.HelpText != nil {
		existingField.HelpText = updates.HelpText
	}
	if updates.DefaultValue != nil {
		existingField.DefaultValue = updates.DefaultValue
	}
	if updates.Options != nil {
		existingField.Options = updates.Options
	}
	if updates.MinLength != nil {
		existingField.MinLength = updates.MinLength
	}
	if updates.MaxLength != nil {
		existingField.MaxLength = updates.MaxLength
	}
	if updates.ReferenceTo != nil {
		existingField.ReferenceTo = updates.ReferenceTo
	}
	if updates.Formula != nil {
		existingField.Formula = updates.Formula
	}
	if updates.ReturnType != nil {
		existingField.ReturnType = updates.ReturnType
	}

	// Generate Field ID (reuse existing)
	fieldID := GenerateFieldID(obj.APIName, fieldAPIName)

	// Delegate to SchemaManager
	if err := ms.schemaMgr.SaveFieldMetadataWithIDs(existingField, obj.ID, fieldID, ms.db); err != nil {
		return fmt.Errorf("failed to update field metadata: %w", err)
	}

	ms.invalidateCacheLocked()
	return nil
}

// DeleteField deletes a field from an object
func (ms *MetadataService) DeleteField(objectAPIName, fieldAPIName string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get the object
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), objectAPIName)
	if err != nil || obj == nil {
		return fmt.Errorf("object '%s' not found", objectAPIName)
	}

	// Find the field
	var existingField *models.FieldMetadata
	for i := range obj.Fields {
		if obj.Fields[i].APIName == fieldAPIName {
			existingField = &obj.Fields[i]
			break
		}
	}
	if existingField == nil {
		return fmt.Errorf("field '%s' not found on object '%s'", fieldAPIName, objectAPIName)
	}

	// Don't allow deleting system fields
	if existingField.IsSystem || existingField.IsNameField {
		return fmt.Errorf("cannot delete system or name field '%s'", fieldAPIName)
	}

	// Delegate to SchemaManager
	if err := ms.schemaMgr.DropColumn(objectAPIName, fieldAPIName); err != nil {
		return fmt.Errorf("failed to drop column: %w", err)
	}

	ms.invalidateCacheLocked()
	return nil
}

// BatchSyncSystemFields batch-upserts multiple system fields (metadata only, no DDL)
func (ms *MetadataService) BatchSyncSystemFields(objectAPIName string, fields []models.FieldMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Get object ID
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), objectAPIName)
	if err != nil || obj == nil {
		return fmt.Errorf("object '%s' not found", objectAPIName)
	}

	// Build batch context for all fields
	batchFields := make([]FieldWithContext, 0, len(fields))
	for _, field := range fields {
		f := field // copy for pointer
		f.IsSystem = true

		batchFields = append(batchFields, FieldWithContext{
			ObjectID: obj.ID,
			FieldID:  GenerateFieldID(objectAPIName, f.APIName),
			Field:    &f,
		})
	}

	// Batch insert/update metadata
	if err := ms.schemaMgr.BatchSaveFieldMetadata(batchFields, ms.db); err != nil {
		return fmt.Errorf("failed to batch sync fields for %s: %w", objectAPIName, err)
	}

	ms.invalidateCacheLocked()
	return nil
}
