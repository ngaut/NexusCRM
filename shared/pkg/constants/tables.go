package constants

import "strings"

// SystemTablePrefix is the prefix for all system tables
const SystemTablePrefix = "_System_"

// Standard Object API Names
const (
	TableAccount = "account"
	TableContact = "contact"
	// Common Fields
	FieldPriority = "priority"
)

// IsSystemTable checks if a table name is a system table
func IsSystemTable(tableName string) bool {
	return strings.HasPrefix(tableName, SystemTablePrefix)
}

// GetSystemFieldNames returns the list of system fields that every table should have.
// Note: id is NOT required since some tables use different primary keys (key_name, composite keys).
func GetSystemFieldNames() []string {
	// Only temporal fields are universally required for operation tracking
	return []string{
		FieldCreatedDate,
		FieldLastModifiedDate,
	}
}

// GetStandardSystemFieldNames returns the full set of system fields for standard/custom objects
func GetStandardSystemFieldNames() []string {
	return []string{
		FieldID,
		FieldCreatedDate,
		FieldLastModifiedDate,
		FieldCreatedByID,
		FieldLastModifiedByID,
		FieldOwnerID,
		FieldIsDeleted,
	}
}

// StandardSystemFields returns the list of standard system field names
// that are present on most business objects.
func StandardSystemFields() []string {
	return []string{
		FieldID,
		FieldOwnerID,
		FieldCreatedDate,
		FieldCreatedByID,
		FieldLastModifiedDate,
		FieldLastModifiedByID,
		FieldIsDeleted,
	}
}

// IsSystemField checks if a field name is a standard system field
func IsSystemField(fieldName string) bool {
	for _, sf := range StandardSystemFields() {
		if sf == fieldName {
			return true
		}
	}
	return false
}

// AuditFields returns the audit field names
func AuditFields() []string {
	return []string{
		FieldCreatedDate,
		FieldCreatedByID,
		FieldLastModifiedDate,
		FieldLastModifiedByID,
	}
}
