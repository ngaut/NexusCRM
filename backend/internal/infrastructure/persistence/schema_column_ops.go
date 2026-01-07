package persistence

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// AddColumn adds a column to the table and registers it
func (r *SchemaRepository) AddColumn(tableName string, col schema.ColumnDefinition) error {
	log.Printf("➕ Adding column %s to table %s", col.Name, tableName)

	// VALIDATION: Table Name
	// System tables (starting with _System_) are exempt from strict snake_case but we generally don't add columns to them dynamically
	if !constants.IsSystemTable(tableName) {
		validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
		if !validName.MatchString(tableName) {
			return fmt.Errorf("invalid table name '%s': must be snake_case", tableName)
		}
	}

	// VALIDATION: Field Definition
	// This calls ValidateFieldDefinition which is still in Service/Manager?
	// ValidateFieldDefinition relies on domain logic. It should be in Manager.
	// But AddColumn here calls it.
	// FIX: Move ValidateFieldDefinition to Repo or pass as dependency?
	// Easier to move ValidateFieldDefinition to Repo since it validates 'col schema.ColumnDefinition'
	// Let's assume r.ValidateFieldDefinition exists (we need to move it or it will fail compilation).
	if err := r.ValidateFieldDefinition(col); err != nil {
		return err
	}

	// 0. IDEMPOTENCY CHECK: Check if column already exists (Zombie/Orphan recovery)
	exists, err := r.checkColumnExists(tableName, col.Name)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if exists {
		log.Printf("⚠️  Orphan column detected: %s.%s exists in DB but missing in metadata. Skipping DDL and adopting column...", tableName, col.Name)
	} else {
		// 1. DDL: ALTER TABLE ADD COLUMN
		ddl := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", tableName, r.buildColumnDDL(col))
		log.Printf("   🏁 Executing DDL: %s", ddl)
		if _, err := r.db.Exec(ddl); err != nil {
			log.Printf("   ❌ DDL execution failed: %v", err)
			return fmt.Errorf("failed to add column to table %s: %w", tableName, err)
		}
		log.Printf("   ✅ DDL execution complete")
	}

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

	// 1.5. DDL: ADD FOREIGN KEY (if applicable)
	if fkDDL != "" {
		log.Printf("   🔗 Adding Foreign Key constraint...")
		if _, err := r.db.Exec(fkDDL); err != nil {
			// Handle "Duplicate constraint" error for idempotency
			if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "already exists") {
				log.Printf("⚠️  FK Constraint %s already exists, skipping...", fmt.Sprintf("fk_%s_%s", tableName, col.Name))
			} else {
				// Only rollback if we actually created the column in this run
				if !exists {
					log.Printf("⚠️ Failed to add FK, rolling back column: %v", err)
					if dropErr := r.DropColumn(tableName, col.Name); dropErr != nil {
						log.Printf("⚠️ Rollback column drop failed: %v", dropErr)
					}
				}
				return fmt.Errorf("failed to add foreign key constraint: %w", err)
			}
		}
	}

	// 2. Register in _System_Field
	if err := r.registerField(tableName, col, r.db); err != nil {
		log.Printf("⚠️  Failed to register field %s.%s: %v. Attempting rollback...", tableName, col.Name, err)

		// COMPENSATION
		if !exists {
			rollbackDDL := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, col.Name)
			if _, rbErr := r.db.Exec(rollbackDDL); rbErr != nil {
				// Critical error: Data vs Metadata inconsistency
				log.Printf("🔥 CRITICAL: Failed to rollback column %s.%s after metadata failure: %v", tableName, col.Name, rbErr)
				return fmt.Errorf("failed to register field AND failed to rollback DDL (critical inconsistency): %w", err)
			}
			log.Printf("   ✅ Rollback successful: Column %s.%s dropped", tableName, col.Name)
		} else {
			log.Printf("⚠️  Skipping rollback for adopted orphan column %s.%s to preserve data.", tableName, col.Name)
		}

		return fmt.Errorf("failed to register field metadata (DDL rolled back: %v): %w", !exists, err)
	}

	log.Printf("   ✅ Column added/registered: %s.%s", tableName, col.Name)
	return nil
}

// EnsureColumn checks if a column exists and adds it if missing
func (r *SchemaRepository) EnsureColumn(tableName string, col schema.ColumnDefinition) error {
	exists, err := r.checkColumnExists(tableName, col.Name)
	if err != nil {
		return err
	}

	if exists {
		// Column exists, no action needed
		return nil
	}

	// Column missing, add it
	log.Printf("⚠️  Column %s.%s missing, adding it...", tableName, col.Name)
	return r.AddColumn(tableName, col)
}

// checkColumnExists queries INFORMATION_SCHEMA to see if a column exists
func (r *SchemaRepository) checkColumnExists(tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND COLUMN_NAME = ?
	`
	var count int
	if err := r.db.QueryRow(query, tableName, columnName).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// DropColumn drops a column from the table and unregisters it
func (r *SchemaRepository) DropColumn(tableName string, columnName string) error {
	log.Printf("➖ Dropping column %s from table %s", columnName, tableName)

	// VALIDATION: Table/Column Name
	if !constants.IsSystemTable(tableName) {
		validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
		if !validName.MatchString(tableName) {
			return fmt.Errorf("invalid table name '%s': must be snake_case", tableName)
		}
		if !validName.MatchString(columnName) {
			return fmt.Errorf("invalid column name '%s': must be snake_case", columnName)
		}
	}

	// 0. IDEMPOTENCY CHECK: Check if column exists
	exists, err := r.checkColumnExists(tableName, columnName)
	if err != nil {
		return fmt.Errorf("failed to check column existence: %w", err)
	}

	if !exists {
		log.Printf("⚠️  Ghost column detected: %s.%s missing from DB but exists in metadata. Skipping DDL and removing metadata...", tableName, columnName)
	} else {
		// 1. DDL: ALTER TABLE DROP COLUMN
		ddl := fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", tableName, columnName)
		if _, err := r.db.Exec(ddl); err != nil {
			return fmt.Errorf("failed to drop column from table %s: %w", tableName, err)
		}
	}

	// 2. Unregister from _System_Field
	fieldID := GenerateFieldID(tableName, columnName)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", constants.TableField)
	if _, err := r.db.Exec(query, fieldID); err != nil {
		log.Printf("⚠️  Warning: Failed to unregister field %s: %v", fieldID, err)
	}

	log.Printf("   ✅ Column dropped: %s.%s", tableName, columnName)
	return nil
}

// registerField registers a single field in _System_Field
func (r *SchemaRepository) registerField(tableName string, col schema.ColumnDefinition, exec Executor) error {
	if exec == nil {
		exec = r.db
	}
	objectID := GenerateObjectID(tableName)
	fieldID := GenerateFieldID(tableName, col.Name)

	// Determine Field Type
	fieldType := r.mapSQLTypeToLogical(col.Type)
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
	isSystem := r.IsSystemColumn(col.Name)

	// Only mark as Required if NOT NULL AND not a system field
	required := !col.Nullable && !isSystem

	// Determine Label
	label := col.Label
	if label == "" {
		// Default to Title Case of snake_case name (e.g. "first_name" -> "First Name")
		label = cases.Title(language.English).String(strings.ReplaceAll(col.Name, "_", " "))
	}

	// Populate FieldMetadata for helper
	field := &models.FieldMetadata{
		APIName:     col.Name,
		Label:       label,
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

	return r.SaveFieldMetadataWithIDs(field, objectID, fieldID, exec)
}
