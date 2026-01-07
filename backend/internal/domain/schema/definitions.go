package schema

// ColumnDefinition represents a single column in a table
type ColumnDefinition struct {
	Name             string   `json:"name"`
	Label            string   `json:"label,omitempty"` // User-friendly label (defaults to Title Cased Name)
	Type             string   `json:"type"`
	LogicalType      string   `json:"logicalType,omitempty"` // Optional: Override logical type (e.g. Password, Picklist)
	PrimaryKey       bool     `json:"primaryKey,omitempty"`
	Unique           bool     `json:"unique,omitempty"`
	Nullable         bool     `json:"nullable,omitempty"`
	Default          string   `json:"default,omitempty"`
	AutoIncrement    bool     `json:"autoIncrement,omitempty"`
	ReferenceTo      []string `json:"referenceTo,omitempty"`
	IsPolymorphic    bool     `json:"isPolymorphic,omitempty"`
	Formula          string   `json:"formula,omitempty"`
	ReturnType       string   `json:"returnType,omitempty"`
	OnDelete         string   `json:"onDelete,omitempty"` // CASCADE, SET NULL, RESTRICT
	IsMasterDetail   bool     `json:"isMasterDetail,omitempty"`
	RelationshipName string   `json:"relationshipName,omitempty"`
	IsNameField      bool     `json:"isNameField,omitempty"`
	Options          []string `json:"options,omitempty"`
	Length           int      `json:"length,omitempty"`
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
	OnDelete   string `json:"onDelete,omitempty"`
	OnUpdate   string `json:"onUpdate,omitempty"`
}

// TableDefinition represents a complete table schema
type TableDefinition struct {
	TableName     string                 `json:"tableName"`
	TableType     string                 `json:"tableType"` // system_core, system_metadata, custom_object
	Category      string                 `json:"category"`  // auth, metadata, crm, etc.
	Label         string                 `json:"label,omitempty"`
	Description   string                 `json:"description"`
	IsManaged     bool                   `json:"isManaged,omitempty"`
	SchemaVersion string                 `json:"schemaVersion,omitempty"`
	Columns       []ColumnDefinition     `json:"columns"`
	Indices       []IndexDefinition      `json:"indices,omitempty"`
	ForeignKeys   []ForeignKeyDefinition `json:"foreignKeys,omitempty"`
}
