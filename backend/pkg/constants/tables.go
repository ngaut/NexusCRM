package constants

import "strings"

// System table names - used throughout the codebase for metadata and system operations.
// These are the internal tables that store platform configuration.
const (
	// Core Metadata Tables
	SystemTablePrefix = "_System_"
	// TableObject, TableField, TableRelationship, TableRecordType defined in z_generated.go
	TableAutoNumber              = "_System_AutoNumber"
	TableTable                   = "_System_Table"
	TableGroup                   = "_System_Group"
	TableGroupMember             = "_System_GroupMember"
	TablePermissionSet           = "_System_PermissionSet"
	TablePermissionSetAssignment = "_System_PermissionSetAssignment"

	// UI & Layout Tables
	// TableLayout defined in z_generated.go
	// TableApp defined in z_generated.go
	TableTheme = "_System_Theme"
	// TableDashboard, TableListView, TableUIComponent, TableSetupPage defined in z_generated.go

	// Security & Permissions Tables
	// TableProfile, TableUser, TableRole defined in z_generated.go
	TableSession = "_System_Session"
	// TableObjectPerms, TableFieldPerms, TableSharingRule defined in z_generated.go
	TableRecordShare = "_System_RecordShare"
	TableTeamMember  = "_System_TeamMember"
	// TableProfileLayout defined in z_generated.go

	// Automation Tables
	// TableFlow defined in z_generated.go
	TableFlowStep     = "_System_FlowStep"
	TableFlowInstance = "_System_FlowInstance"
	// TableAction, TableValidation defined in z_generated.go
	TableFieldDependency  = "_System_FieldDependency"
	TableApprovalProcess  = "_System_ApprovalProcess"
	TableApprovalWorkItem = "_System_ApprovalWorkItem"

	// System Operation Tables
	// TableRecycleBin, TableLog, TableRecent, TableConfig defined in z_generated.go
	TableComment      = "_System_Comment"
	TableNotification = "_System_Notification"
	TableAuditLog     = "_System_AuditLog"
	TableOutboxEvent  = "_System_OutboxEvent"
)

func IsSystemTable(tableName string) bool {
	return strings.HasPrefix(tableName, SystemTablePrefix)
}

// GetSystemFieldNames returns the list of system fields that every table should have.
// Note: id is NOT required since some tables use different primary keys (key_name, composite keys).
func GetSystemFieldNames() []string {
	// Only temporal fields are universally required for operation tracking
	return []string{
		"created_date",
		"last_modified_date",
	}
}

// GetStandardSystemFieldNames returns the full set of system fields for standard/custom objects
func GetStandardSystemFieldNames() []string {
	return []string{
		"id",
		"created_date",
		"last_modified_date",
		"created_by_id",
		"last_modified_by_id",
		"owner_id",
		"is_deleted",
	}
}
