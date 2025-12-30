package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/models"
	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/pkg/constants"
)

// AddColumn adds a column to the table and registers it
func (sm *SchemaManager) AddColumn(tableName string, col schema.ColumnDefinition) error {
	log.Printf("➕ Adding column %s to table %s", col.Name, tableName)

	// VALIDATION
	if err := sm.ValidateFieldDefinition(col); err != nil {
		return err
	}

	// 1. DDL: ALTER TABLE ADD COLUMN
	ddl := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", tableName, sm.buildColumnDDL(col))

	// If it's a reference field, we handle the Foreign Key in a separate statement
	var fkDDL string
	if col.ReferenceTo != "" {
		fkName := fmt.Sprintf("fk_%s_%s", tableName, col.Name)
		fkDDL = fmt.Sprintf("ALTER TABLE `%s` ADD CONSTRAINT `%s` FOREIGN KEY (`%s`) REFERENCES `%s` (`id`)",
			tableName, fkName, col.Name, col.ReferenceTo)

		if col.OnDelete != "" {
			fkDDL += fmt.Sprintf(" ON DELETE %s", col.OnDelete)
		}
	}

	log.Printf("   🏁 Executing DDL: %s", ddl)
	if _, err := sm.db.Exec(ddl); err != nil {
		log.Printf("   ❌ DDL execution failed: %v", err)
		return fmt.Errorf("failed to add column to table %s: %w", tableName, err)
	}
	log.Printf("   ✅ DDL execution complete")

	// 1.5. DDL: ADD FOREIGN KEY (if applicable)
	if fkDDL != "" {
		log.Printf("   🔗 Adding Foreign Key constraint...")
		if _, err := sm.db.Exec(fkDDL); err != nil {
			// Rollback: Drop the column we just added
			log.Printf("⚠️ Failed to add FK, rolling back column: %v", err)
			if dropErr := sm.DropColumn(tableName, col.Name); dropErr != nil {
				log.Printf("⚠️ Rollback column drop failed: %v", dropErr)
			}
			return fmt.Errorf("failed to add foreign key constraint (rolled back column): %w", err)
		}
	}

	// 2. Register in _System_Field
	if err := sm.registerField(tableName, col, sm.db); err != nil {
		log.Printf("⚠️  Failed to register field %s.%s: %v. Attempting rollback...", tableName, col.Name, err)

		// COMPENSATION: Attempt to drop the column to maintain consistency
		rollbackDDL := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, col.Name)
		if _, rbErr := sm.db.Exec(rollbackDDL); rbErr != nil {
			// Critical error: Data vs Metadata inconsistency
			log.Printf("🔥 CRITICAL: Failed to rollback column %s.%s after metadata failure: %v", tableName, col.Name, rbErr)
			return fmt.Errorf("failed to register field AND failed to rollback DDL (critical inconsistency): %w", err)
		}

		log.Printf("   ✅ Rollback successful: Column %s.%s dropped", tableName, col.Name)
		return fmt.Errorf("failed to register field metadata (DDL rolled back): %w", err)
	}

	log.Printf("   ✅ Column added: %s.%s", tableName, col.Name)
	return nil
}

// EnsureColumn checks if a column exists and adds it if missing
func (sm *SchemaManager) EnsureColumn(tableName string, col schema.ColumnDefinition) error {
	// Query INFORMATION_SCHEMA to check existence
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND COLUMN_NAME = ?
	`
	var count int
	if err := sm.db.QueryRow(query, tableName, col.Name).Scan(&count); err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if count > 0 {
		// Column exists, no action needed
		return nil
	}

	// Column missing, add it
	log.Printf("⚠️  Column %s.%s missing, adding it...", tableName, col.Name)
	return sm.AddColumn(tableName, col)
}

// DropColumn drops a column from the table and unregisters it
func (sm *SchemaManager) DropColumn(tableName string, columnName string) error {
	log.Printf("➖ Dropping column %s from table %s", columnName, tableName)

	// 1. DDL: ALTER TABLE DROP COLUMN
	ddl := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, columnName)
	if _, err := sm.db.Exec(ddl); err != nil {
		return fmt.Errorf("failed to drop column from table %s: %w", tableName, err)
	}

	// 2. Unregister from _System_Field
	fieldID := fmt.Sprintf("fld_%s_%s", tableName, columnName)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableField)
	if _, err := sm.db.Exec(query, fieldID); err != nil {
		log.Printf("⚠️  Warning: Failed to unregister field %s: %v", fieldID, err)
	}

	log.Printf("   ✅ Column dropped: %s.%s", tableName, columnName)
	return nil
}

// registerField registers a single field in _System_Field
func (sm *SchemaManager) registerField(tableName string, col schema.ColumnDefinition, exec Executor) error {
	if exec == nil {
		exec = sm.db
	}
	objectID := GenerateObjectID(tableName)
	fieldID := fmt.Sprintf("fld_%s_%s", tableName, col.Name)

	// Determine Field Type
	fieldType := sm.mapSQLTypeToLogical(col.Type)
	if col.LogicalType != "" {
		fieldType = col.LogicalType
	}

	isNameField := false
	if strings.EqualFold(col.Name, constants.FieldName) {
		isNameField = true
	}
	if col.IsNameField {
		isNameField = true
	}

	// Detect system columns
	isSystem := sm.IsSystemColumn(col.Name)

	// Only mark as Required if NOT NULL AND not a system field
	required := !col.Nullable && !isSystem

	// Populate FieldMetadata for helper
	field := &models.FieldMetadata{
		APIName:     col.Name,
		Label:       col.Name,
		Type:        models.FieldType(fieldType),
		Required:    required,
		Unique:      col.Unique,
		IsSystem:    isSystem,
		IsNameField: isNameField,
	}

	// DESIGN ASSUMPTION: AllReferences contains the full list of referenced objects for polymorphic lookups.
	if len(col.AllReferences) > 0 {
		field.ReferenceTo = col.AllReferences
	} else {
		field.ReferenceTo = WrapStringToSlice(col.ReferenceTo)
	}

	// Persist Options
	if len(col.Options) > 0 {
		field.Options = col.Options
	}

	field.IsPolymorphic = len(field.ReferenceTo) > 1
	field.IsMasterDetail = col.IsMasterDetail

	if col.OnDelete != "" {
		dr := models.DeleteRule(col.OnDelete)
		field.DeleteRule = &dr
	}

	if col.RelationshipName != "" {
		field.RelationshipName = &col.RelationshipName
	}

	if col.Default != "" {
		field.DefaultValue = &col.Default
	}

	if col.Formula != "" {
		field.Formula = &col.Formula
	}
	if col.ReturnType != "" {
		rt := models.FieldType(col.ReturnType)
		field.ReturnType = &rt
	}

	return sm.SaveFieldMetadataWithIDs(field, objectID, fieldID, exec)
}
