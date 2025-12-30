package models

// SObject represents a generic CRM object record (map of field names to values)
type SObject map[string]interface{}

// FieldType alias
type FieldType = string

// ObjectMetadata describes the schema of a CRM object
type ObjectMetadata struct {
	APIName     string          `json:"api_name"`
	Label       string          `json:"label"`
	PluralLabel string          `json:"plural_label"`
	Description *string         `json:"description,omitempty"`
	IsCustom    bool            `json:"is_custom,omitempty"`
	Fields      []FieldMetadata `json:"fields"`
}

// FieldMetadata describes a single field on an object
type FieldMetadata struct {
	APIName     string    `json:"api_name"`
	Label       string    `json:"label"`
	Type        FieldType `json:"type"` // Changed from data_type to match backend
	Required    bool      `json:"required"`
	Description *string   `json:"description,omitempty"`
	Options     []string  `json:"options,omitempty"`
	ReferenceTo []string  `json:"reference_to,omitempty"`
}

// UserSession represents the authenticated user context
type UserSession struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	ProfileID     string `json:"profile_id"`
	IsSystemAdmin bool   `json:"is_system_admin"`
}

type QueryRequest struct {
	ObjectAPIName string `json:"object_api_name"`
	FilterExpr    string `json:"filter_expr,omitempty"` // Formula expression for filtering
	SortField     string `json:"sort_field,omitempty"`
	SortDirection string `json:"sort_direction,omitempty"`
	Limit         int    `json:"limit,omitempty"`
}

// DashboardWidget represents a widget configuration for dashboards
type DashboardWidget struct {
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`                   // list, chart, metric, sql_chart, etc.
	Object      string                 `json:"object,omitempty"`       // Target object for data
	Filter      string                 `json:"filter,omitempty"`       // Filter expression
	Columns     []string               `json:"columns,omitempty"`      // Fields to display
	ChartType   string                 `json:"chart_type,omitempty"`   // pie, bar, line, etc.
	GroupBy     string                 `json:"group_by,omitempty"`     // Field to group by
	AggField    string                 `json:"agg_field,omitempty"`    // Field to aggregate
	AggFunction string                 `json:"agg_function,omitempty"` // count, sum, avg, etc.
	Size        string                 `json:"size,omitempty"`         // small, medium, large
	SQL         string                 `json:"sql,omitempty"`          // SQL query for sql_chart type
	Config      map[string]interface{} `json:"config,omitempty"`       // Additional configuration
}

// DashboardCreate is the request payload for creating a dashboard
type DashboardCreate struct {
	Name        string            `json:"name"`
	Label       string            `json:"label,omitempty"`
	Description string            `json:"description,omitempty"`
	Layout      string            `json:"layout,omitempty"` // two-column, grid, etc.
	Widgets     []DashboardWidget `json:"widgets"`
}
