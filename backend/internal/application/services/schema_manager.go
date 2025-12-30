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

// CreateTableFromDefinition creates a table from a declarative definition
func (sm *SchemaManager) CreateTableFromDefinition(ctx context.Context, def schema.TableDefinition) error {
	// 1. Create Physical Table
	if err := sm.CreatePhysicalTable(ctx, def); err != nil {
		return err
	}

	// TRANSACTION: Register Metadata
	// We wrap metadata registration in a transaction to ensure atomicity.
	// If registration fails, we must manually compensate by dropping the physical table.
	tx, err := sm.db.Begin()
	if err != nil {
		// Attempt to cleanup
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 1. Register in _System_Table registry
	if err := sm.registerTable(def, tx); err != nil {
		_ = tx.Rollback()
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to register table %s: %w", def.TableName, err)
	}

	// 2. Register in _System_Object and _System_Field for metadata-driven operations
	if err := sm.RegisterObjectMetadata(def, tx); err != nil {
		_ = tx.Rollback()
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to register object metadata %s: %w", def.TableName, err)
	}

	if err := tx.Commit(); err != nil {
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to commit metadata transaction: %w", err)
	}

	log.Printf("   ✅ Table created and registered: %s", def.TableName)
	return nil
}

// CreateTableWithBatchMetadata creates a table using batch metadata registration (optimized)
// This is faster than CreateTableFromDefinition for tables with many fields
func (sm *SchemaManager) CreateTableWithBatchMetadata(ctx context.Context, def schema.TableDefinition, objectMeta *models.ObjectMetadata) error {
	// 1. Create Physical Table (DDL only)
	if err := sm.CreatePhysicalTable(ctx, def); err != nil {
		return err
	}

	// 2. Batch register table in _System_Table
	if err := sm.BatchRegisterTables([]schema.TableDefinition{def}, sm.db); err != nil {
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to register table %s: %w", def.TableName, err)
	}

	// 3. Batch register object in _System_Object
	if err := sm.BatchSaveObjectMetadata([]*models.ObjectMetadata{objectMeta}, sm.db); err != nil {
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to register object %s: %w", def.TableName, err)
	}

	// 4. Batch register all fields in _System_Field
	batchFields := make([]FieldWithContext, 0, len(def.Columns))
	for _, col := range def.Columns {
		batchFields = append(batchFields, sm.PrepareFieldForBatch(def.TableName, col))
	}

	if err := sm.BatchSaveFieldMetadata(batchFields, sm.db); err != nil {
		if _, dropErr := sm.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS `%s`", def.TableName)); dropErr != nil {
			log.Printf("⚠️ Failed to cleanup table %s: %v", def.TableName, dropErr)
		}
		return fmt.Errorf("failed to register fields for %s: %w", def.TableName, err)
	}

	log.Printf("   ✅ Table created and registered: %s", def.TableName)
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

	log.Printf("   ✅ Table dropped and metadata cleaned: %s", tableName)
	return nil
}

// RegisterObjectMetadata registers the table definition as a logical object in _System_Object and _System_Field
// This promotes the table to a "First Class" object in the metadata system
func (sm *SchemaManager) RegisterObjectMetadata(def schema.TableDefinition, exec Executor) error {
	if exec == nil {
		exec = sm.db
	}
	// 1. Construct Object Metadata
	objectID := GenerateObjectID(def.TableName)
	isCustom := constants.TableType(def.TableType) == constants.TableTypeCustomObject

	// Determine Label (use description or table name if not provided)
	label := def.Description
	if label == "" {
		label = def.TableName
	}

	description := def.Description

	obj := &models.ObjectMetadata{
		ID:           objectID,
		APIName:      def.TableName,
		Label:        label,
		PluralLabel:  def.TableName + "s", // Simple pluralization
		Description:  &description,
		IsCustom:     isCustom,
		SharingModel: models.SharingModel(constants.SharingModelPublicReadWrite), // Default for system objects
	}

	// 2. Upsert Object (reuse Batch method)
	if err := sm.BatchSaveObjectMetadata([]*models.ObjectMetadata{obj}, exec); err != nil {
		return fmt.Errorf("failed to register object %s: %w", def.TableName, err)
	}

	// 3. Register Fields in _System_Field (reuse Batch method)
	batchFields := make([]FieldWithContext, 0, len(def.Columns))
	for _, col := range def.Columns {
		batchFields = append(batchFields, sm.PrepareFieldForBatch(def.TableName, col))
	}
	if err := sm.BatchSaveFieldMetadata(batchFields, exec); err != nil {
		return fmt.Errorf("failed to register fields for %s: %w", def.TableName, err)
	}

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
	sqlType := fieldtypes.GetSQLType(fieldType)
	if sqlType == "" {
		// Default fallback for unknown types
		return "VARCHAR(255)"
	}
	return sqlType
}

// System column functions are in schema_system_columns.go:
// - GetStandardSystemColumns, GetStandardFieldMetadata

// ValidateSchema checks if a table matches its definition
func (sm *SchemaManager) ValidateSchema(tableName string) error {
	// NOTE: Schema drift detection not implemented. Would require querying INFORMATION_SCHEMA
	// to compare actual database schema with expected schema from definitions. This is useful
	// for detecting manual schema changes or migration issues. For now, relies on clean bootstrap.
	// Future: Implement SchemaService.DetectDrift() to compare and report differences.
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
	switch field.Type {
	case string(constants.FieldTypeLookup):
		// User assumption: Lookups must specific a target
		if field.ReferenceTo == "" && len(field.AllReferences) == 0 {
			return fmt.Errorf("lookup field '%s' must have a valid 'reference_to' target", field.Name)
		}
	case string(constants.FieldTypePicklist):
		// User assumption: Picklists must have options
		if len(field.Options) == 0 {
			// Note: We allow empty options temporarily if loaded from partial metadata?
			// But creating a new field via schema without options should fail.
			// Let's enforce it strictly "Faile Fast".
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
