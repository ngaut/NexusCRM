package services

import (
	"context"
	"fmt"
	"log"

	"github.com/nexuscrm/backend/pkg/errors"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// ==================== Schema CRUD Methods ====================

// CreateSchema creates a new custom object schema and physical table
func (ms *MetadataService) CreateSchema(ctx context.Context, schema *models.ObjectMetadata) error {
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
	existing, err := ms.repo.GetSchemaByAPIName(ctx, schema.APIName)
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
	if err := ms.schemaMgr.CreateTableWithStrictMetadata(ctx, def, schema); err != nil {
		return fmt.Errorf("failed to create schema via SchemaManager: %w", err)
	}

	// ==================== AUTO-GENERATE DEFAULT LAYOUT ====================
	defaultLayout := ms.GenerateDefaultLayout(schema)

	// Persist Layout to _System_Layout
	// Persist Layout to _System_Layout via Repo
	if err := ms.repo.UpsertLayout(ctx, &defaultLayout); err != nil {
		log.Printf("⚠️ Failed to auto-create default layout for %s: %v", schema.APIName, err)
	} else {
		log.Printf("✅ Auto-created default layout for %s", schema.APIName)
	}

	ms.invalidateCacheLocked()
	return nil
}

// CreateSchemaOptimized creates a new object schema using batch metadata registration
// This is faster than CreateSchema for objects with many fields
func (ms *MetadataService) CreateSchemaOptimized(ctx context.Context, schema *models.ObjectMetadata) error {
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
	existing, err := ms.repo.GetSchemaByAPIName(ctx, schema.APIName)
	if err == nil && existing != nil {
		return errors.NewConflictError("Object Metadata", "api_name", schema.APIName)
	}

	// Create table with STRICT metadata registration (Fails on Unique Constraint)
	// Note: PrepareTableDefinition updates schema with defaults, so we pass it back
	if err := ms.schemaMgr.CreateTableWithStrictMetadata(ctx, def, schema); err != nil {
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
func (ms *MetadataService) UpdateSchema(ctx context.Context, apiName string, updates *models.ObjectMetadata) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	obj, err := ms.repo.GetSchemaByAPIName(ctx, apiName)
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
	if updates.ThemeColor != nil {
		obj.ThemeColor = updates.ThemeColor
	}

	// Use helper to persist changes
	if err := ms.schemaMgr.SaveObjectMetadata(obj, nil); err != nil {
		return fmt.Errorf("failed to update object: %w", err)
	}

	ms.invalidateCacheLocked()
	return nil
}

// DeleteSchema deletes a custom object schema and drops its table
func (ms *MetadataService) DeleteSchema(ctx context.Context, apiName string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Check if exists
	obj, err := ms.repo.GetSchemaByAPIName(ctx, apiName)
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
