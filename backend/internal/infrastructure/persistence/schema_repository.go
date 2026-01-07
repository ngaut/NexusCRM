package persistence

import (
	"context"
	"database/sql"
	"strings"

	"github.com/nexuscrm/shared/pkg/constants"
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
// It handles both Logical Types (from constants) and Physical SQL Types (for system tables)
func (r *SchemaRepository) MapFieldTypeToSQL(fieldType string) string {
	// 1. Check Logical Types (using shared constants)
	switch constants.SchemaFieldType(fieldType) {
	case constants.FieldTypeText:
		return "VARCHAR(255)"
	case constants.FieldTypeTextArea:
		return "TEXT"
	case constants.FieldTypeLongTextArea:
		return "LONGTEXT"
	case constants.FieldTypeRichText:
		return "LONGTEXT"
	case constants.FieldTypeNumber:
		return "DECIMAL(18,6)"
	case constants.FieldTypeCurrency:
		return "DECIMAL(18,2)"
	case constants.FieldTypePercent:
		return "DECIMAL(5,2)"
	case constants.FieldTypeBoolean:
		return "BOOLEAN"
	case constants.FieldTypeDate:
		return "DATE"
	case constants.FieldTypeDateTime:
		return "DATETIME"
	case constants.FieldTypeEmail:
		return "VARCHAR(255)"
	case constants.FieldTypePhone:
		return "VARCHAR(50)"
	case constants.FieldTypeURL:
		return "VARCHAR(255)"
	case constants.FieldTypePicklist:
		return "VARCHAR(255)"
	case constants.FieldTypeMultiPicklist:
		return "JSON" // Stored as JSON array
	case constants.FieldTypeLookup, constants.FieldTypeMasterDetail:
		return "VARCHAR(36)" // UUID
	case constants.FieldTypeAutoNumber:
		return "VARCHAR(255)" // Usually string based format
	case constants.FieldTypeFormula:
		return "VARCHAR(255)" // Depends on return type, handled by buildColumnDDL
	case constants.FieldTypePassword, constants.FieldTypeEncryptedString:
		return "VARCHAR(255)"
	case constants.FieldTypeJSON:
		return "JSON"
	}

	// 2. Check Raw SQL Types (Passthrough for System Tables)
	// We allow specific raw types that match what we use in system_tables.json
	upper := strings.ToUpper(fieldType)
	switch upper {
	case "INT", "INTEGER", "TINYINT", "TINYINT(1)", "BIGINT", "SMALLINT":
		return fieldType // Keep original casing/precision if needed, usually uppercase
	case "DATETIME", "TIMESTAMP", "DATE":
		return upper
	case "TEXT", "MEDIUMTEXT", "LONGTEXT", "JSON":
		return upper
	case "BOOLEAN", "BOOL":
		return "BOOLEAN"
	case "VARCHAR(255)", "VARCHAR(36)", "VARCHAR(50)", "CHAR(36)":
		return fieldType
	}

	// Default fallback
	return "VARCHAR(255)"
}
