package schema

// ColumnDefinition represents a single column in a table
type ColumnDefinition struct {
	Name             string   `json:"name"`
	Type             string   `json:"type"`
	LogicalType      string   `json:"logical_type,omitempty"` // Optional: Override logical type (e.g. Password, Picklist)
	PrimaryKey       bool     `json:"primary_key,omitempty"`
	Unique           bool     `json:"unique,omitempty"`
	Nullable         bool     `json:"nullable,omitempty"`
	Default          string   `json:"default,omitempty"`
	AutoIncrement    bool     `json:"auto_increment,omitempty"`
	ReferenceTo      string   `json:"reference_to,omitempty"`
	AllReferences    []string `json:"all_references,omitempty"`
	Formula          string   `json:"formula,omitempty"`
	ReturnType       string   `json:"return_type,omitempty"`
	OnDelete         string   `json:"on_delete,omitempty"` // CASCADE, SET NULL, RESTRICT
	IsMasterDetail   bool     `json:"is_master_detail,omitempty"`
	RelationshipName string   `json:"relationship_name,omitempty"`
	IsNameField      bool     `json:"is_name_field,omitempty"`
	Options          []string `json:"options,omitempty"`
}

// IndexDefinition represents an index on a table
type IndexDefinition struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

// ForeignKeyDefinition represents a foreign key constraint
type ForeignKeyDefinition struct {
	Column     string `json:"column"`
	References string `json:"references"` // format: "tableName(columnName)"
	OnDelete   string `json:"on_delete,omitempty"`
	OnUpdate   string `json:"on_update,omitempty"`
}

// TableDefinition represents a complete table schema
type TableDefinition struct {
	TableName   string                 `json:"table_name"`
	TableType   string                 `json:"table_type"` // system_core, system_metadata, custom_object
	Category    string                 `json:"category"`   // auth, metadata, crm, etc.
	Description string                 `json:"description"`
	Columns     []ColumnDefinition     `json:"columns"`
	Indices     []IndexDefinition      `json:"indices,omitempty"`
	ForeignKeys []ForeignKeyDefinition `json:"foreign_keys,omitempty"`
}
