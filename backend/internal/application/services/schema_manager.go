package services

import (
	"context"
	"database/sql"

	"github.com/nexuscrm/backend/internal/domain/schema"
	"github.com/nexuscrm/backend/internal/infrastructure/persistence"
	"github.com/nexuscrm/shared/pkg/models"
)

// SchemaManager handles all table creation and schema operations
type SchemaManager struct {
	repo *persistence.SchemaRepository
}

// NewSchemaManager creates a new schema manager
func NewSchemaManager(repo *persistence.SchemaRepository) *SchemaManager {
	return &SchemaManager{
		repo: repo,
	}
}

// CreatePhysicalTable creates the table structure without registering metadata
func (sm *SchemaManager) CreatePhysicalTable(ctx context.Context, def schema.TableDefinition) error {
	return sm.repo.CreatePhysicalTable(ctx, def)
}

// CreateTableWithStrictMetadata creates a table ensuring metadata uniqueness (Strict Insert)
func (sm *SchemaManager) CreateTableWithStrictMetadata(ctx context.Context, def schema.TableDefinition, objectMeta *models.ObjectMetadata) (err error) {
	return sm.repo.CreateTableWithStrictMetadata(ctx, def, objectMeta)
}

// DDL generation functions are in schema_ddl.go:
// - buildColumnDDL
// - convertFormulaToSQL
// - buildForeignKeyDDL
// - buildIndexDDL

// BatchCreatePhysicalTables performs parallel DDL creation and then batch registers in _System_Table
func (sm *SchemaManager) BatchCreatePhysicalTables(ctx context.Context, defs []schema.TableDefinition) error {
	return sm.repo.BatchCreatePhysicalTables(ctx, defs)
}

// DropTable drops a table and removes it from the registry
func (sm *SchemaManager) DropTable(tableName string) error {
	return sm.repo.DropTable(tableName)
}

// Column operation functions are in schema_column_ops.go:
// - AddColumn, EnsureColumn, DropColumn, registerField

// Methods moved to SchemaRepository
// System column functions are in schema_system_columns.go:
// - GetStandardSystemColumns, GetStandardFieldMetadata

// ValidateSchema checks if a table matches its expected definition
func (sm *SchemaManager) ValidateSchema(tableName string) error {
	return sm.repo.ValidateSchema(tableName)
}

// ValidateFieldDefinition validates the field schema against core assertions
func (sm *SchemaManager) ValidateFieldDefinition(field schema.ColumnDefinition) error {
	return sm.repo.ValidateFieldDefinition(field)
}

// ValidateSchemaRegistry compares the registry against actual DB tables
func (sm *SchemaManager) ValidateSchemaRegistry() (*persistence.SchemaHealth, error) {
	return sm.repo.ValidateSchemaRegistry()
}

// =========================================================================================
// Proxies for moved methods (Backward Compatibility)
// =========================================================================================

// Type Aliases for compatibility
type FieldWithContext = persistence.FieldWithContext
type TableRegistryItem = persistence.TableRegistryItem
type SchemaHealth = persistence.SchemaHealth

// BatchRegisterTables registers multiple tables in _System_Table registry
func (sm *SchemaManager) BatchRegisterTables(defs []schema.TableDefinition, tx *sql.Tx) error {
	return sm.repo.BatchRegisterTables(defs, tx)
}

// AddColumn adds a column to the table and registers it
func (sm *SchemaManager) AddColumn(tableName string, col schema.ColumnDefinition) error {
	return sm.repo.AddColumn(tableName, col)
}

// EnsureColumn checks if a column exists and adds it if missing
func (sm *SchemaManager) EnsureColumn(tableName string, col schema.ColumnDefinition) error {
	return sm.repo.EnsureColumn(tableName, col)
}

// DropColumn drops a column from the table and removes metadata
func (sm *SchemaManager) DropColumn(tableName string, colName string) error {
	return sm.repo.DropColumn(tableName, colName)
}

// IsSystemColumn returns true for columns that are automatically populated
func (sm *SchemaManager) IsSystemColumn(name string) bool {
	return sm.repo.IsSystemColumn(name)
}

// GetStandardSystemColumns returns the default columns for every custom object
func (sm *SchemaManager) GetStandardSystemColumns() []schema.ColumnDefinition {
	return sm.repo.GetStandardSystemColumns()
}

// GetStandardFieldMetadata returns field metadata for standard system columns
func (sm *SchemaManager) GetStandardFieldMetadata() []models.FieldMetadata {
	return sm.repo.GetStandardFieldMetadata()
}

// SaveObjectMetadata upserts object metadata into _System_Object
func (sm *SchemaManager) SaveObjectMetadata(obj *models.ObjectMetadata, tx *sql.Tx) error {
	return sm.repo.SaveObjectMetadata(obj, tx)
}

// InsertObjectMetadata inserts object metadata (Strict)
func (sm *SchemaManager) InsertObjectMetadata(obj *models.ObjectMetadata, tx *sql.Tx) error {
	return sm.repo.InsertObjectMetadata(obj, tx)
}

// BatchSaveObjectMetadata inserts multiple objects in a single statement
func (sm *SchemaManager) BatchSaveObjectMetadata(objs []*models.ObjectMetadata, tx *sql.Tx) error {
	return sm.repo.BatchSaveObjectMetadata(objs, tx)
}

// SaveFieldMetadataWithIDs upserts field metadata with explicit IDs
func (sm *SchemaManager) SaveFieldMetadataWithIDs(field *models.FieldMetadata, objectID string, fieldID string, tx *sql.Tx) error {
	return sm.repo.SaveFieldMetadataWithIDs(field, objectID, fieldID, tx)
}

// BatchSaveFieldMetadata inserts multiple fields in a single statement
func (sm *SchemaManager) BatchSaveFieldMetadata(fields []FieldWithContext, tx *sql.Tx) error {
	return sm.repo.BatchSaveFieldMetadata(fields, tx)
}

// PrepareFieldForBatch converts a column definition to FieldWithContext
func (sm *SchemaManager) PrepareFieldForBatch(tableName string, col schema.ColumnDefinition) FieldWithContext {
	return sm.repo.PrepareFieldForBatch(tableName, col)
}

// GetTableRegistry retrieves all registered tables
func (sm *SchemaManager) GetTableRegistry() ([]*TableRegistryItem, error) {
	return sm.repo.GetTableRegistry()
}

// MapFieldTypeToSQL converts logical field types to SQL column types
func (sm *SchemaManager) MapFieldTypeToSQL(fieldType string) string {
	return sm.repo.MapFieldTypeToSQL(fieldType)
}

// ValidateFormula validates a formula expression syntax
func (sm *SchemaManager) ValidateFormula(formulaStr string, env map[string]interface{}) error {
	return sm.repo.ValidateFormula(formulaStr, env)
}
