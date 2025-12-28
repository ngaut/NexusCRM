package constants

// System field names - standard fields present on all/most objects.
// These are the snake_case API names used in storage and SQL.
const (
	// Primary Fields defined in z_generated.go
	// FieldID, FieldName

	// Ownership and Security defined in z_generated.go
	// FieldOwnerID

	// Audit Fields defined in z_generated.go
	// FieldCreatedDate, FieldCreatedByID, FieldLastModifiedDate, FieldLastModifiedByID

	// Soft Delete defined in z_generated.go
	// FieldIsDeleted

	// User Fields
	FieldEmail         = "email"
	FieldUsername      = "username"
	FieldPassword      = "password"
	FieldFirstName     = "first_name"
	FieldLastName      = "last_name"
	FieldProfileID     = "profile_id"
	FieldRoleID        = "role_id"
	FieldIsActive      = "is_active"
	FieldLastLoginDate = "last_login_date"

	// Object/Field Metadata
	FieldAPIName     = "api_name"
	FieldLabel       = "label"
	FieldPluralLabel = "plural_label"
	FieldDescription = "description"
	FieldMetaType    = "type"
	FieldObjectID    = "object_id"
	FieldReferenceTo = "reference_to"
	FieldIsCustom    = "is_custom"
	FieldIsSystem    = "is_system"
	FieldIsRequired  = "required"
	FieldIsUnique    = "unique"

	// Flow/Action Fields
	FieldTriggerObject = "trigger_object"
	FieldTriggerType   = "trigger_type"
	FieldCondition     = "condition"
	FieldConfig        = "config"
	FieldStatus        = "status"
	FieldSortOrder     = "sort_order"

	// Recycle Bin Fields
	FieldRecordID      = "record_id"
	FieldObjectAPIName = "object_api_name"
	FieldRecordName    = "record_name"
	FieldDeletedBy     = "deleted_by"
	FieldDeletedDate   = "deleted_date"

	// Recent Items Fields
	FieldUserID    = "user_id"
	FieldTimestamp = "timestamp"

	// Config Fields
	FieldKeyName  = "key_name"
	FieldValue    = "value"
	FieldIsSecret = "is_secret"

	// Log Fields
	FieldLevel   = "level"
	FieldSource  = "source"
	FieldMessage = "message"
	FieldDetails = "details"

	// Session Fields
	FieldToken        = "token"
	FieldExpiresAt    = "expires_at"
	FieldLastActivity = "last_activity"
	FieldIPAddress    = "ip_address"
	FieldUserAgent    = "user_agent"
	FieldIsRevoked    = "is_revoked"

	// Metadata Fields
	FieldTableName     = "table_name"
	FieldTableType     = "table_type"
	FieldCategory      = "category"
	FieldIsManaged     = "is_managed"
	FieldSchemaVersion = "schema_version"
	FieldCreatedBy     = "created_by"
)

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
