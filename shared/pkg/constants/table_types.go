package constants

// TableType defines the type of table (system vs custom)
type TableType string

// Table type constants - used for schema classification
const (
	TableTypeCustomObject   TableType = "custom_object"
	TableTypeSystemMetadata TableType = "system_metadata"
	TableTypeSystemCore     TableType = "system_core"
	TableTypeSystemData     TableType = "system_data"
	TableTypeSystemJunction TableType = "system_junction"
)
