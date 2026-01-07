package persistence

import (
	"context"
	"database/sql"
	"strings"
)

// SchemaRepository handles direct database schema operations (DDL) and system registry updates
type SchemaRepository struct {
	db *sql.DB
}

// NewSchemaRepository creates a new SchemaRepository
func NewSchemaRepository(db *sql.DB) *SchemaRepository {
	return &SchemaRepository{
		db: db,
	}
}

// Executor interface for db/tx flexibility
type Executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// IsSystemColumn returns true for columns that are automatically populated
// by the server or database (e.g., timestamps, IDs, ownership fields)
func (r *SchemaRepository) IsSystemColumn(name string) bool {
	// Common system columns
	switch name {
	case "id", "created_date", "last_modified_date", "created_by_id", "last_modified_by_id", "owner_id", "is_deleted":
		return true
	}
	return false
}

// mapSQLTypeToLogical converts SQL types to system logical types
func (r *SchemaRepository) mapSQLTypeToLogical(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)
	if strings.HasPrefix(sqlType, "VARCHAR") || strings.HasPrefix(sqlType, "TEXT") || strings.HasPrefix(sqlType, "CHAR") {
		return "Text"
	}
	if strings.HasPrefix(sqlType, "BOOL") || strings.HasPrefix(sqlType, "TINYINT(1)") {
		return "Checkbox"
	}
	if strings.HasPrefix(sqlType, "INT") || strings.HasPrefix(sqlType, "BIGINT") || strings.HasPrefix(sqlType, "TINYINT") {
		return "Number"
	}
	if strings.HasPrefix(sqlType, "DECIMAL") || strings.HasPrefix(sqlType, "FLOAT") || strings.HasPrefix(sqlType, "DOUBLE") {
		return "Number"
	}
	if strings.HasPrefix(sqlType, "DATETIME") || strings.HasPrefix(sqlType, "TIMESTAMP") {
		return "DateTime"
	}
	if strings.HasPrefix(sqlType, "DATE") {
		return "Date"
	}
	if strings.HasPrefix(sqlType, "JSON") {
		return "JSON" // Special handling
	}
	return "Text" // Default fallback
}

// MapFieldTypeToSQL converts logical field types to SQL column types
// This is used by MetadataService to prepare TableDefinition
// Now uses the centralized fieldtypes registry loaded from shared/constants/fieldTypes.json
func (r *SchemaRepository) MapFieldTypeToSQL(fieldType string) string {
	// Simplified mapping for now, should ideally load from registry
	switch fieldType {
	case "Text":
		return "VARCHAR(255)"
	case "TextArea":
		return "TEXT"
	case "LongTextArea":
		return "LONGTEXT"
	case "Number":
		return "DECIMAL(18,6)"
	case "Currency":
		return "DECIMAL(18,2)"
	case "Percent":
		return "DECIMAL(5,2)"
	case "Checkbox":
		return "BOOLEAN"
	case "Date":
		return "DATE"
	case "DateTime":
		return "DATETIME"
	case "Email":
		return "VARCHAR(255)"
	case "Phone":
		return "VARCHAR(50)"
	case "URL":
		return "VARCHAR(255)"
	case "Picklist":
		return "VARCHAR(255)"
	case "MultiPicklist":
		return "JSON" // Stored as JSON array
	case "Lookup":
		return "VARCHAR(36)" // UUID
	case "MasterDetail":
		return "VARCHAR(36)" // UUID
	case "AutoNumber":
		return "VARCHAR(255)" // Usually string based format
	case "Formula":
		return "VARCHAR(255)" // Depends on return type, handled by buildColumnDDL
	case "EncryptedString":
		return "VARCHAR(255)"
	case "RichText":
		return "LONGTEXT"
	case "JSON":
		return "JSON"
	default:
		return "VARCHAR(255)"
	}
}
