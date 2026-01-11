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
	case constants.FieldID, constants.FieldCreatedDate, constants.FieldLastModifiedDate, constants.FieldCreatedByID, constants.FieldLastModifiedByID, constants.FieldOwnerID, constants.FieldIsDeleted:
		return true
	}
	return false
}

// mapSQLTypeToLogical converts SQL types to system logical types
func (r *SchemaRepository) mapSQLTypeToLogical(sqlType string) string {
	sqlType = strings.ToUpper(sqlType)
	if strings.HasPrefix(sqlType, SQLTypeVarchar) || strings.HasPrefix(sqlType, SQLTypeText) || strings.HasPrefix(sqlType, SQLTypeChar) {
		return string(constants.FieldTypeText)
	}
	if strings.HasPrefix(sqlType, SQLTypeBool) || strings.HasPrefix(sqlType, SQLTypeTinyInt1) {
		return string(constants.FieldTypeBoolean)
	}
	if strings.HasPrefix(sqlType, SQLTypeInt) || strings.HasPrefix(sqlType, SQLTypeBigInt) || strings.HasPrefix(sqlType, SQLTypeTinyInt) {
		return string(constants.FieldTypeNumber)
	}
	if strings.HasPrefix(sqlType, SQLTypeDecimal) || strings.HasPrefix(sqlType, SQLTypeFloat) || strings.HasPrefix(sqlType, SQLTypeDouble) {
		return string(constants.FieldTypeNumber)
	}
	if strings.HasPrefix(sqlType, SQLTypeDateTime) || strings.HasPrefix(sqlType, SQLTypeTimestamp) {
		return string(constants.FieldTypeDateTime)
	}
	if strings.HasPrefix(sqlType, SQLTypeDate) {
		return string(constants.FieldTypeDate)
	}
	if strings.HasPrefix(sqlType, SQLTypeJSON) {
		return string(constants.FieldTypeJSON) // Special handling
	}
	return string(constants.FieldTypeText) // Default fallback
}

// MapFieldTypeToSQL converts logical field types to SQL column types
// This is used by MetadataService to prepare TableDefinition
// It handles both Logical Types (from constants) and Physical SQL Types (for system tables)
func (r *SchemaRepository) MapFieldTypeToSQL(fieldType string) string {
	// 1. Check Logical Types (using shared constants)
	switch constants.SchemaFieldType(fieldType) {
	case constants.FieldTypeText:
		return SQLTypeVarchar255
	case constants.FieldTypeTextArea:
		return SQLTypeText
	case constants.FieldTypeLongTextArea:
		return SQLTypeLongText
	case constants.FieldTypeRichText:
		return SQLTypeLongText
	case constants.FieldTypeNumber:
		return SQLTypeDecimal18_6
	case constants.FieldTypeCurrency:
		return SQLTypeDecimal18_2
	case constants.FieldTypePercent:
		return SQLTypeDecimal5_2
	case constants.FieldTypeBoolean:
		return SQLTypeBoolean
	case constants.FieldTypeDate:
		return SQLTypeDate
	case constants.FieldTypeDateTime:
		return SQLTypeDateTime
	case constants.FieldTypeEmail:
		return SQLTypeVarchar255
	case constants.FieldTypePhone:
		return SQLTypeVarchar50
	case constants.FieldTypeURL:
		return SQLTypeText
	case constants.FieldTypePicklist:
		return SQLTypeVarchar255
	case constants.FieldTypeMultiPicklist:
		return SQLTypeJSON // Stored as JSON array
	case constants.FieldTypeLookup, constants.FieldTypeMasterDetail:
		return SQLTypeVarchar36 // UUID
	case constants.FieldTypeAutoNumber:
		return SQLTypeVarchar255 // Usually string based format
	case constants.FieldTypeFormula:
		return SQLTypeVarchar255 // Depends on return type, handled by buildColumnDDL
	case constants.FieldTypePassword, constants.FieldTypeEncryptedString:
		return SQLTypeText
	case constants.FieldTypeJSON:
		return SQLTypeJSON
	}

	// 2. Check Raw SQL Types (Passthrough for System Tables)
	// We allow specific raw types that match what we use in system_tables.json
	upper := strings.ToUpper(fieldType)
	switch upper {
	case SQLTypeInt, SQLTypeInteger, SQLTypeTinyInt, SQLTypeTinyInt1, SQLTypeBigInt, SQLTypeSmallInt:
		return fieldType // Keep original casing/precision if needed, usually uppercase
	case SQLTypeDateTime, SQLTypeTimestamp, SQLTypeDate:
		return upper
	case SQLTypeText, "MEDIUMTEXT", "LONGTEXT", SQLTypeJSON:
		return upper
	case SQLTypeBoolean, SQLTypeBool:
		return SQLTypeBoolean
	case "VARCHAR(255)", "VARCHAR(36)", "VARCHAR(50)", "CHAR(36)":
		return fieldType
	}

	// Default fallback
	return SQLTypeVarchar255
}
