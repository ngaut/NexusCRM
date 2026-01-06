package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/pkg/fieldtypes"
	"github.com/nexuscrm/shared/pkg/constants"
	"github.com/nexuscrm/shared/pkg/models"
)

// SchemaManager handles all table creation and schema operations
type SchemaManager struct {
	db *sql.DB
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(db *sql.DB) *SchemaManager {
	return &SchemaManager{db: db}
}

// CreatePhysicalTable creates the table structure without registering metadata
func (sm *SchemaManager) CreatePhysicalTable(ctx context.Context, def schema.TableDefinition) error {
	log.Printf("📐 Creating table: %s", def.TableName)

	// VALIDATION: Table Name
	// System tables (starting with _System_) are exempt from strict snake_case
	if !constants.IsSystemTable(def.TableName) {
		validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
		if !validName.MatchString(def.TableName) {
			return fmt.Errorf("table name '%s' must be snake_case (lowercase, alphanumeric, underscores)", def.TableName)
		}
	}

	// Build CREATE TABLE statement with indexes inline
	var ddl strings.Builder
	ddl.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", def.TableName))

	// Add columns
	for i, col := range def.Columns {
		// VALIDATION: Fail fast if schema assumptions are violated
		if err := sm.ValidateFieldDefinition(col); err != nil {
			return fmt.Errorf("invalid column definition for '%s': %w", col.Name, err)
		}

		ddl.WriteString("  ")
		ddl.WriteString(sm.buildColumnDDL(col))
		// Always add comma if there are more columns, indexes, or foreign keys
		if i < len(def.Columns)-1 || len(def.Indices) > 0 || len(def.ForeignKeys) > 0 {
			ddl.WriteString(",")
		}
		ddl.WriteString("\n")
	}

	// Add indexes inline (KEY or UNIQUE KEY)
	for i, idx := range def.Indices {
		ddl.WriteString("  ")
		ddl.WriteString(sm.buildIndexDDL(def.TableName, idx))
		if i < len(def.Indices)-1 || len(def.ForeignKeys) > 0 {
			ddl.WriteString(",")
		}
		ddl.WriteString("\n")
	}

	// Add foreign keys
	for i, fk := range def.ForeignKeys {
		ddl.WriteString("  ")
		ddl.WriteString(sm.buildForeignKeyDDL(fk))
		if i < len(def.ForeignKeys)-1 {
			ddl.WriteString(",")
		}
		ddl.WriteString("\n")
	}
	ddl.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")

	// Execute DDL using a dedicated connection to ensure SET FOREIGN_KEY_CHECKS works

	conn, err := sm.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Disable foreign key checks for this DDL session
	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS=0"); err != nil {
		log.Printf("⚠️ Failed to disable FK checks: %v", err)
	}

	log.Printf("📝 Executing DDL for %s", def.TableName)
	if _, err := conn.ExecContext(ctx, ddl.String()); err != nil {
		log.Printf("❌ Failed to create table %s: %v", def.TableName, err)
		return fmt.Errorf("failed to create table %s: %w", def.TableName, err)
	}
	log.Printf("✅ DDL executed successfully for %s", def.TableName)

	return nil
}

// CreateTableWithStrictMetadata creates a table ensuring metadata uniqueness (Strict Insert)
func (sm *SchemaManager) CreateTableWithStrictMetadata(ctx context.Context, def schema.TableDefinition, objectMeta *models.ObjectMetadata) (err error) {
	// 1. Create Physical Table
	if err = sm.CreatePhysicalTable(ctx, def); err != nil {
		return err
	}

	// COMPENSATION: Ensure table is dropped if any subsequent step fails
	// We use the named return 'err' to determine if cleanup is needed
	defer func() {
		if err != nil {
			log.Printf("⚠️ An error occurred during registration. Rolling back table creation: %s", def.TableName)
			if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
				log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
			}
		}
	}()

	// TRANSACTION: Register Metadata
	tx, err := sm.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Ensure rollback on panic or error (safe to call after Commit)
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Register in _System_Table
	if err = sm.registerTable(def, tx); err != nil {
		return fmt.Errorf("failed to register table %s: %w", def.TableName, err)
	}

	// 2. Register Object Metadata (Strict)
	// We use InsertObjectMetadata instead of BatchSaveObjectMetadata to fail on uniqueness
	if err = sm.InsertObjectMetadata(objectMeta, tx); err != nil {
		return fmt.Errorf("failed to register object (strict) %s: %w", def.TableName, err)
	}

	// 3. Register Fields (Batch is OK here as fields belong to new object)
	batchFields := make([]FieldWithContext, 0, len(def.Columns))
	for _, col := range def.Columns {
		batchFields = append(batchFields, sm.PrepareFieldForBatch(def.TableName, col))
	}

	if err = sm.BatchSaveFieldMetadata(batchFields, tx); err != nil {
		return fmt.Errorf("failed to register fields for %s: %w", def.TableName, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit metadata transaction: %w", err)
	}

	log.Printf("   ✅ Table created and registered (strict): %s", def.TableName)
	return nil
}

// DDL generation functions are in schema_ddl.go:
// - buildColumnDDL
// - convertFormulaToSQL
// - buildForeignKeyDDL
// - buildIndexDDL

// BatchCreatePhysicalTables performs parallel DDL creation and then batch registers in _System_Table
func (sm *SchemaManager) BatchCreatePhysicalTables(ctx context.Context, defs []schema.TableDefinition) error {
	// 1. Parallel DDL Execution
	// TiDB handles concurrent DDL well. We limit concurrency to avoid overwhelming the connection pool.
	sem := make(chan struct{}, 10) // Limit to 10 concurrent DDLs
	var wg sync.WaitGroup
	errChan := make(chan error, len(defs))

	// Track successfully created tables for compensation
	var createdTablesMu sync.Mutex
	var createdTables []string

	log.Printf("🚀 Starting parallel DDL for %d tables...", len(defs))

	for _, def := range defs {
		wg.Add(1)
		go func(d schema.TableDefinition) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			// Call CreatePhysicalTable (DDL only)
			if err := sm.CreatePhysicalTable(ctx, d); err != nil {
				errChan <- fmt.Errorf("failed to create table %s: %w", d.TableName, err)
			} else {
				createdTablesMu.Lock()
				createdTables = append(createdTables, d.TableName)
				createdTablesMu.Unlock()
			}
		}(def)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	if len(errChan) > 0 {
		// COMPENSATION: Drop successfully created tables
		firstErr := <-errChan
		log.Printf("❌ Batch DDL failed (%v). Rolling back %d created tables...", firstErr, len(createdTables))

		for _, tableName := range createdTables {
			if err := sm.DropTable(tableName); err != nil {
				log.Printf("⚠️ Failed to cleanup table %s during rollback: %v", tableName, err)
			}
		}

		return firstErr // Return first error
	}

	// 2. Batch Register in _System_Table
	log.Printf("📦 Batch registering %d tables in _System_Table...", len(defs))
	if err := sm.BatchRegisterTables(defs, sm.db); err != nil {
		// COMPENSATION: Drop all physical tables if metadata registration fails
		log.Printf("❌ Batch registration failed (%v). Rolling back %d created tables...", err, len(createdTables))
		for _, tableName := range createdTables {
			if dropErr := sm.DropTable(tableName); dropErr != nil {
				log.Printf("⚠️ Failed to cleanup table %s during rollback: %v", tableName, dropErr)
			}
		}
		return fmt.Errorf("failed to batch register tables: %w", err)
	}

	return nil
}

// registerTable registers the table in _System_Table registry
func (sm *SchemaManager) registerTable(def schema.TableDefinition, exec Executor) error {
	if exec == nil {
		exec = sm.db
	}
	return sm.BatchRegisterTables([]schema.TableDefinition{def}, exec)
}

// BatchRegisterTables registers multiple tables in _System_Table registry in a single statement
func (sm *SchemaManager) BatchRegisterTables(defs []schema.TableDefinition, exec Executor) error {
	if len(defs) == 0 {
		return nil
	}
	if exec == nil {
		exec = sm.db
	}

	var valuePlaceholders []string
	var args []interface{}

	for _, def := range defs {
		id := GenerateTableID(def.TableName)
		valuePlaceholders = append(valuePlaceholders, "(?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())")
		args = append(args,
			id,
			def.TableName,
			def.TableType,
			def.Category,
			def.Description,
			def.IsManaged,
			def.SchemaVersion, // Use version from definition
			"bootstrap",       // Created by
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, table_name, table_type, category, description, is_managed, schema_version, created_by, created_date, last_modified_date)
		VALUES %s
		ON DUPLICATE KEY UPDATE last_modified_date = NOW()
	`, constants.TableTable, strings.Join(valuePlaceholders, ", "))

	_, err := sm.db.Exec(query, args...)
	return err
}

// DropTable drops a table and removes it from the registry
func (sm *SchemaManager) DropTable(tableName string) error {
	log.Printf("🔥 Dropping table: %s", tableName)

	// Drop the table
	if _, err := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)); err != nil {
		return fmt.Errorf("failed to drop table %s: %w", tableName, err)
	}

	// Unregister from _System_Table
	if _, err := sm.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE table_name = ?", constants.TableTable), tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to unregister table %s: %v", tableName, err)
	}

	// UNREGISTER from _System_Object and _System_Field
	// Deleting the Object metadata should cascade to fields if configured,
	// but we explicitly delete fields first to ensure clean removal.

	// Delete fields for this object
	fieldDeleteQuery := fmt.Sprintf("DELETE FROM %s WHERE object_id IN (SELECT id FROM %s WHERE api_name = ?)", constants.TableField, constants.TableObject)
	if _, err := sm.db.Exec(fieldDeleteQuery, tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to delete fields for object %s: %v", tableName, err)
	}

	// Delete object
	if _, err := sm.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE api_name = ?", constants.TableObject), tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to delete object metadata %s: %v", tableName, err)
	}

	// Delete AutoNumber metadata
	if _, err := sm.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE LOWER(object_api_name) = LOWER(?)", constants.TableAutoNumber), tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to delete auto-number metadata for %s: %v", tableName, err)
	}

	// Delete Object Permissions
	if _, err := sm.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE object_api_name = ?", constants.TableObjectPerms), tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to delete object permissions for %s: %v", tableName, err)
	}

	// Delete Field Permissions
	if _, err := sm.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE object_api_name = ?", constants.TableFieldPerms), tableName); err != nil {
		log.Printf("⚠️  Warning: Failed to delete field permissions for %s: %v", tableName, err)
	}

	log.Printf("   ✅ Table dropped and metadata cleaned: %s", tableName)
	return nil
}

// Column operation functions are in schema_column_ops.go:
// - AddColumn, EnsureColumn, DropColumn, registerField

// IsSystemColumn returns true for columns that are automatically populated
// by the server or database (e.g., timestamps, IDs, ownership fields)
func (sm *SchemaManager) IsSystemColumn(name string) bool {
	// Normalize to lowercase for comparison
	lower := strings.ToLower(name)
	return constants.IsSystemField(lower)
}

// mapSQLTypeToLogical converts SQL types to system logical types
func (sm *SchemaManager) mapSQLTypeToLogical(sqlType string) string {
	upper := strings.ToUpper(sqlType)

	// Check exact overrides or specific types first
	if strings.Contains(upper, "BOOLEAN") || strings.Contains(upper, "TINYINT(1)") {
		return string(constants.FieldTypeBoolean)
	}

	if strings.Contains(upper, "VARCHAR") || strings.Contains(upper, "TEXT") || strings.Contains(upper, "CHAR") {
		return string(constants.FieldTypeText)
	}
	if strings.Contains(upper, "INT") || strings.Contains(upper, "DECIMAL") || strings.Contains(upper, "FLOAT") || strings.Contains(upper, "DOUBLE") {
		return string(constants.FieldTypeNumber)
	}
	if strings.Contains(upper, "DATETIME") || strings.Contains(upper, "TIMESTAMP") {
		return string(constants.FieldTypeDateTime)
	}
	if strings.Contains(upper, "DATE") {
		return string(constants.FieldTypeDate)
	}
	if strings.Contains(upper, string(constants.FieldTypeJSON)) {
		return string(constants.FieldTypeJSON)
	}
	return string(constants.FieldTypeText) // Default
}

// MapFieldTypeToSQL converts logical field types to SQL column types
// This is used by MetadataService to prepare TableDefinition
// Now uses the centralized fieldtypes registry loaded from shared/constants/fieldTypes.json
func (sm *SchemaManager) MapFieldTypeToSQL(fieldType string) string {
	// Special handling for AutoNumber which is logically a string in storage
	if strings.EqualFold(fieldType, string(constants.FieldTypeAutoNumber)) {
		return "VARCHAR(255)"
	}

	sqlType := fieldtypes.GetSQLType(fieldType)
	if sqlType == "" {
		// Default fallback: assumes the type is already a valid SQL type (e.g. DATETIME, TINYINT, VARCHAR)
		// This supports system tables that define raw SQL types in their JSON schema.
		return fieldType
	}
	return sqlType
}

// System column functions are in schema_system_columns.go:
// - GetStandardSystemColumns, GetStandardFieldMetadata

// ValidateSchema checks if a table matches its expected definition by comparing
// the actual database schema (from INFORMATION_SCHEMA) with the registered metadata.
// Returns an error if drift is detected, nil if schema matches or table doesn't exist.
func (sm *SchemaManager) ValidateSchema(tableName string) error {
	// Get expected columns from _System_Field metadata
	expectedQuery := fmt.Sprintf(`
		SELECT f.api_name, f.type 
		FROM %s f 
		JOIN %s o ON f.object_id = o.id 
		WHERE o.api_name = ?
	`, constants.TableField, constants.TableObject)

	expectedRows, err := sm.db.Query(expectedQuery, tableName)
	if err != nil {
		return fmt.Errorf("failed to query expected schema: %w", err)
	}
	defer func() { _ = expectedRows.Close() }()

	expectedFields := make(map[string]string) // field_name -> type
	for expectedRows.Next() {
		var fieldName, fieldType string
		if err := expectedRows.Scan(&fieldName, &fieldType); err != nil {
			continue
		}
		expectedFields[fieldName] = fieldType
	}

	// If no metadata registered, skip validation (table might be new or system-internal)
	if len(expectedFields) == 0 {
		return nil
	}

	// Get actual columns from INFORMATION_SCHEMA
	actualQuery := `
		SELECT COLUMN_NAME, DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
	`
	actualRows, err := sm.db.Query(actualQuery, tableName)
	if err != nil {
		return fmt.Errorf("failed to query actual schema: %w", err)
	}
	defer func() { _ = actualRows.Close() }()

	actualFields := make(map[string]string)
	for actualRows.Next() {
		var columnName, dataType string
		if err := actualRows.Scan(&columnName, &dataType); err != nil {
			continue
		}
		actualFields[columnName] = dataType
	}

	// Compare: Find missing columns in database
	var missingColumns []string
	for expectedField := range expectedFields {
		if _, exists := actualFields[expectedField]; !exists {
			missingColumns = append(missingColumns, expectedField)
		}
	}

	if len(missingColumns) > 0 {
		return fmt.Errorf("schema drift detected for %s: missing columns %v", tableName, missingColumns)
	}

	return nil
}

// Metadata persistence functions are in schema_metadata_persist.go:
// - FieldWithContext, SaveObjectMetadata, SaveFieldMetadataWithIDs
// - BatchSaveObjectMetadata, BatchSaveFieldMetadata, PrepareFieldForBatch
// - TableRegistryItem, GetTableRegistry

// SchemaHealth represents the health status of the schema registry
type SchemaHealth struct {
	Status         string   `json:"status"`
	ExpectedCount  int      `json:"expectedCount"`
	ActualCount    int      `json:"actualCount"`
	MissingTables  []string `json:"missingTables"`
	RegistryHealth string   `json:"registryHealth"`
}

// ValidateFieldDefinition validates the field schema against core assertions
func (sm *SchemaManager) ValidateFieldDefinition(field schema.ColumnDefinition) error {
	// 1. Naming Convention: snake_case
	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validName.MatchString(field.Name) {
		return fmt.Errorf("field name '%s' must be snake_case (lowercase, alphanumeric, underscores)", field.Name)
	}

	// 2. Type-Specific Assertions
	// Use LogicalType for validation to catch metadata issues (e.g. Picklist options)
	// regardless of the underlying SQL storage type (e.g. VARCHAR)
	checkType := field.LogicalType
	if checkType == "" {
		checkType = field.Type
	}

	switch checkType {
	case string(constants.FieldTypeLookup):
		// User assumption: Lookups must specific a target
		if field.ReferenceTo == "" && len(field.AllReferences) == 0 {
			return fmt.Errorf("lookup field '%s' must have a valid 'reference_to' target", field.Name)
		}
	case string(constants.FieldTypePicklist):
		// User assumption: Picklists must have options
		if len(field.Options) == 0 {
			// Fail Fast: Reject metadata that maps to a Picklist but has no options.
			return fmt.Errorf("picklist field '%s' must have at least one option", field.Name)
		}
	case string(constants.FieldTypeFormula):
		if field.Formula == "" {
			return fmt.Errorf("formula field '%s' must have a formula expression", field.Name)
		}
		if field.ReturnType == "" {
			return fmt.Errorf("formula field '%s' must have a valid return_type", field.Name)
		}
	}

	return nil
}

// ValidateSchemaRegistry compares the registry against actual DB tables
func (sm *SchemaManager) ValidateSchemaRegistry() (*SchemaHealth, error) {
	// Get expected tables
	expectedRows, err := sm.db.Query(fmt.Sprintf("SELECT %s FROM %s WHERE %s = 1", constants.FieldTableName, constants.TableTable, constants.FieldIsManaged))
	if err != nil {
		return nil, err
	}
	defer func() { _ = expectedRows.Close() }()

	expected := make(map[string]bool)
	for expectedRows.Next() {
		var tableName string
		if err := expectedRows.Scan(&tableName); err != nil {
			continue
		}
		expected[tableName] = true
	}

	// Get actual tables
	actualRows, err := sm.db.Query(`
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE()
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = actualRows.Close() }()

	actual := make(map[string]bool)
	for actualRows.Next() {
		var tableName string
		if err := actualRows.Scan(&tableName); err != nil {
			continue
		}
		actual[tableName] = true
	}

	// Find missing tables
	var missing []string
	for exp := range expected {
		if !actual[exp] {
			missing = append(missing, exp)
		}
	}

	status := "healthy"
	if len(missing) > 0 {
		status = "unhealthy"
	}

	return &SchemaHealth{
		Status:         status,
		ExpectedCount:  len(expected),
		ActualCount:    len(actual),
		MissingTables:  missing,
		RegistryHealth: "operational",
	}, nil
}
