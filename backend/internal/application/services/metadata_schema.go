package services

import (
	"context"
	"fmt"
	"log"

	domainSchema "github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Schema CRUD Methods ====================

// CreateSchema creates a new custom object schema and physical table
func (ms *MetadataService) CreateSchema(schema *models.ObjectMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Capture original labels before any processing/defaults
	originalPlural := schema.PluralLabel
	originalLabel := schema.Label

	// Validate Object Metadata
	if ms.validationSvc != nil {
		if err := ms.validationSvc.ValidateObjectMetadata(schema); err != nil {
			return err
		}
	}

	// Prepare Table Definition (Validation, Enrichment, Mapping)
	def, _, err := ms.PrepareTableDefinition(schema)
	if err != nil {
		return err
	}

	// Check if metadata exists in DB
	existing, err := ms.repo.GetSchemaByAPIName(context.Background(), schema.APIName)
	if err == nil && existing != nil {
		return errors.NewConflictError("Object Metadata", "api_name", schema.APIName)
	}

	// Restore original labels for strict insert (if provided/valid)
	// This ensures we persist exactly what the user provided, overriding any "Humanization" done by PrepareTableDefinition
	if originalLabel != "" {
		schema.Label = originalLabel
	}
	if originalPlural != "" {
		schema.PluralLabel = originalPlural
	}

	// Create Table (Strict Mode: Fails if metadata conflicts)
	if err := ms.schemaMgr.CreateTableWithStrictMetadata(context.Background(), def, schema); err != nil {
		return fmt.Errorf("failed to create schema via SchemaManager: %w", err)
	}

	// ==================== AUTO-GENERATE DEFAULT LAYOUT ====================
	defaultLayout := ms.GenerateDefaultLayout(schema)

	// Persist Layout to _System_Layout
	// Persist Layout to _System_Layout via Repo
	if err := ms.repo.UpsertLayout(context.Background(), &defaultLayout); err != nil {
		log.Printf("⚠️ Failed to auto-create default layout for %s: %v", schema.APIName, err)
	} else {
		log.Printf("✅ Auto-created default layout for %s", schema.APIName)
	}

	ms.invalidateCacheLocked()
	return nil
}

// CreateSchemaOptimized creates a new object schema using batch metadata registration
// This is faster than CreateSchema for objects with many fields
func (ms *MetadataService) CreateSchemaOptimized(schema *models.ObjectMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Validate Object Metadata
	if ms.validationSvc != nil {
		if err := ms.validationSvc.ValidateObjectMetadata(schema); err != nil {
			return err
		}
	}

	// Prepare Table Definition
	def, _, err := ms.PrepareTableDefinition(schema)
	if err != nil {
		return err
	}

	// Check if metadata exists in DB
	existing, err := ms.repo.GetSchemaByAPIName(context.Background(), schema.APIName)
	if err == nil && existing != nil {
		return errors.NewConflictError("Object Metadata", "api_name", schema.APIName)
	}

	// Create table with STRICT metadata registration (Fails on Unique Constraint)
	// Note: PrepareTableDefinition updates schema with defaults, so we pass it back
	if err := ms.schemaMgr.CreateTableWithStrictMetadata(context.Background(), def, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Auto-generate default layout (same as CreateSchema)
	defaultLayout := ms.GenerateDefaultLayout(schema)
	if err := ms.repo.UpsertLayout(context.Background(), &defaultLayout); err != nil {
		log.Printf("⚠️ Failed to insert default layout for %s: %v", schema.APIName, err)
	}

	ms.invalidateCacheLocked()
	return nil
}

// InvalidateCache/invalidateCacheLocked moved to metadata_service.go

// UpdateSchema updates an existing object schema
func (ms *MetadataService) UpdateSchema(apiName string, updates *models.ObjectMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), apiName)
	if err != nil || obj == nil {
		return fmt.Errorf("object with API name '%s' not found", apiName)
	}

	// Prepare updates by modifying the existing object in-memory
	if updates.Label != "" {
		obj.Label = updates.Label
	}
	if updates.PluralLabel != "" {
		obj.PluralLabel = updates.PluralLabel
	}
	if updates.Description != nil {
		obj.Description = updates.Description
	}
	if updates.PathField != nil {
		obj.PathField = updates.PathField
	}
	if len(updates.ListFields) > 0 {
		obj.ListFields = updates.ListFields
	}
	if updates.Icon != "" {
		obj.Icon = updates.Icon
	}
	if updates.AppID != nil {
		obj.AppID = updates.AppID
	}
	if updates.SharingModel != "" {
		obj.SharingModel = updates.SharingModel
	}

	// Use helper to persist changes
	if err := ms.schemaMgr.SaveObjectMetadata(obj, ms.db); err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	ms.invalidateCacheLocked()
	return nil
}

// DeleteSchema deletes a custom object schema and drops its table
func (ms *MetadataService) DeleteSchema(apiName string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	obj, err := ms.repo.GetSchemaByAPIName(context.Background(), apiName)
	if err != nil || obj == nil {
		return fmt.Errorf("object with API name '%s' not found", apiName)
	}

	// Physical table drop and Unregister
	if err := ms.schemaMgr.DropTable(apiName); err != nil {
		return fmt.Errorf("failed to drop table and schema: %w", err)
	}

	// Delete from _System_Object and _System_Field is handled by DropTable internally
	ms.invalidateCacheLocked()
	return nil
}

// Field CRUD methods are in metadata_field_crud.go:
// - CreateField, UpdateField, DeleteField, BatchSyncSystemFields

// EnsureDefaultListView creates a default "All" list view if none exists
func (ms *MetadataService) EnsureDefaultListView(objectAPIName string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	count, err := ms.repo.CountListViews(context.Background(), objectAPIName)
	if err != nil {
		return fmt.Errorf("failed to check list views: %w", err)
	}

	if count > 0 {
		return nil
	}

	id := GenerateID()
	label := "All"
	filterExpr := ""
	// Default fields: Name, CreatedDate, Owner
	fields := []string{constants.FieldName, constants.FieldCreatedDate, constants.FieldOwnerID}

	view := &models.ListView{
		ID:            id,
		ObjectAPIName: objectAPIName,
		Label:         label,
		FilterExpr:    filterExpr,
		Fields:        fields,
	}

	if err := ms.repo.CreateListView(context.Background(), view); err != nil {
		return fmt.Errorf("failed to insert default list view: %w", err)
	}
	log.Printf("✅ Auto-created default list view for %s", objectAPIName)
	return nil
}

// BatchCreateSchemas performs massive batch creation of multiple schemas using the "Super Batch" strategy
// 1. Parallel DDL for all tables
// 2. Single batch insert for _System_Table
// 3. Single batch insert for _System_Object
// 4. Single batch insert for _System_Field
// 5. Batch insert for default layouts
func (ms *MetadataService) BatchCreateSchemas(schemas []models.ObjectMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	log.Printf("🚀 Starting Super Batch creation for %d objects...", len(schemas))

	// 1. Prepare Data
	tableDefs := make([]domainSchema.TableDefinition, 0, len(schemas))
	allObjects := make([]*models.ObjectMetadata, 0, len(schemas))
	allFields := make([]FieldWithContext, 0)
	layouts := make([]models.PageLayout, 0, len(schemas))

	for i := range schemas {
		schema := &schemas[i] // Pointer to original

		// Prepare Table Definition (Validation, Enrichment, Mapping)
		def, batchFields, err := ms.PrepareTableDefinition(schema)
		if err != nil {
			return err
		}

		tableDefs = append(tableDefs, def)
		allObjects = append(allObjects, schema) // pointer, schema was modified by PrepareTableDefinition (e.g. ID)
		allFields = append(allFields, batchFields...)

		// Prepare Default Layout
		layout := ms.GenerateDefaultLayout(schema)
		layouts = append(layouts, layout)
	}

	// 2. Execute Batch Operations
	// A. Physical Tables & Registry (Parallel DDL + Batch Table Insert)
	if err := ms.schemaMgr.BatchCreatePhysicalTables(context.Background(), tableDefs); err != nil {
		return fmt.Errorf("batch physical table creation failed: %w", err)
	}

	// B. Batch Object Metadata
	log.Printf("📦 Batch inserting %d objects in _System_Object...", len(allObjects))
	if err := ms.schemaMgr.BatchSaveObjectMetadata(allObjects, ms.db); err != nil {
		return fmt.Errorf("batch object metadata save failed: %w", err)
	}

	// C. Batch Field Metadata
	log.Printf("📦 Batch inserting %d fields in _System_Field...", len(allFields))
	if err := ms.schemaMgr.BatchSaveFieldMetadata(allFields, ms.db); err != nil {
		return fmt.Errorf("batch field metadata save failed: %w", err)
	}

	// D. Batch Layouts
	log.Printf("📦 Batch inserting %d layouts in _System_Layout...", len(layouts))
	if len(layouts) > 0 {
		ctx := context.Background()
		// Convert []models.PageLayout to []*models.PageLayout for BatchUpsertLayouts
		layoutPointers := make([]*models.PageLayout, len(layouts))
		for i := range layouts {
			layoutPointers[i] = &layouts[i]
		}
		if err := ms.repo.BatchUpsertLayouts(ctx, layoutPointers); err != nil {
			log.Printf("⚠️ Failed to batch insert layouts: %v", err)
		}
	}

	// E. Batch Default Permissions (for custom objects)
	if ms.permissionSvc != nil {
		profiles := []string{constants.ProfileSystemAdmin, constants.ProfileStandardUser}
		for _, schema := range schemas {
			if !schema.IsCustom {
				continue
			}
			for _, field := range schema.Fields {
				for _, profileID := range profiles {
					fieldPerm := models.FieldPermission{
						ProfileID:     &profileID,
						ObjectAPIName: schema.APIName,
						FieldAPIName:  field.APIName,
						Readable:      true,
						Editable:      true,
					}
					if err := ms.permissionSvc.UpdateFieldPermission(fieldPerm); err != nil {
						log.Printf("⚠️ Failed to grant permission for field %s: %v", field.APIName, err)
					}

					// Also grant to the type column if polymorphic
					if field.IsPolymorphic {
						typeFieldPerm := fieldPerm
						typeFieldPerm.FieldAPIName = GetPolymorphicTypeColumnName(field.APIName)
						if err := ms.permissionSvc.UpdateFieldPermission(typeFieldPerm); err != nil {
							log.Printf("⚠️ Failed to grant permission for polymorphic type field: %v", err)
						}
					}
				}
			}
		}
	}

	ms.invalidateCacheLocked()
	return nil
}
