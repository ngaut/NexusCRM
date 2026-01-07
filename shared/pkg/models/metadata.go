package models

import (
	"time"

	"github.com/nexuscrm/shared/pkg/constants"
)

// FieldType is defined in pkg/constants
type FieldType = constants.SchemaFieldType

// SharingModel is defined in pkg/constants
type SharingModel = constants.SharingModel

// DeleteRule is defined in pkg/constants
type DeleteRule = constants.DeleteRule

// RollupConfig represents rollup summary field configuration
type RollupConfig struct {
	SummaryObject     string  `json:"summary_object"`
	SummaryField      string  `json:"summary_field"`
	RelationshipField string  `json:"relationship_field"` // The lookup field on the child object
	CalcType          string  `json:"calc_type"`          // COUNT, SUM, MIN, MAX, AVG
	Filter            *string `json:"filter,omitempty"`
}

// FieldMetadata represents field-level metadata
type FieldMetadata struct {
	ID                 string              `json:"id,omitempty"`
	APIName            string              `json:"api_name"`
	Label              string              `json:"label"`
	Type               FieldType           `json:"type"`
	Required           bool                `json:"required,omitempty"`
	Unique             bool                `json:"unique,omitempty"`
	IsNameField        bool                `json:"is_name_field,omitempty"`
	Options            []string            `json:"options,omitempty"`
	ReferenceTo        []string            `json:"reference_to,omitempty"`   // Supports polymorphic (multiple objects)
	IsPolymorphic      bool                `json:"is_polymorphic,omitempty"` // True if len(ReferenceTo) > 1
	DeleteRule         *DeleteRule         `json:"delete_rule,omitempty"`
	IsSystem           bool                `json:"is_system,omitempty"`
	Formula            *string             `json:"formula,omitempty"`
	ReturnType         *FieldType          `json:"return_type,omitempty"`
	DefaultValue       *string             `json:"default_value,omitempty"`
	HelpText           *string             `json:"help_text,omitempty"`
	TrackHistory       bool                `json:"track_history,omitempty"`
	MinValue           *float64            `json:"min_value,omitempty"`
	MaxValue           *float64            `json:"max_value,omitempty"`
	MinLength          *int                `json:"min_length,omitempty"`
	MaxLength          *int                `json:"max_length,omitempty"`
	Regex              *string             `json:"regex,omitempty"`
	RegexMessage       *string             `json:"regex_message,omitempty"`
	Validator          *string             `json:"validator,omitempty"`
	ValidatorConfig    *string             `json:"validator_config,omitempty"` // JSON config for the validator
	ControllingField   *string             `json:"controlling_field,omitempty"`
	PicklistDependency map[string][]string `json:"picklist_dependency,omitempty"`
	RollupConfig       *RollupConfig       `json:"rollup_config,omitempty"`
	IsMasterDetail     bool                `json:"is_master_detail,omitempty"`
	RelationshipName   *string             `json:"relationship_name,omitempty"`
}

// ObjectMetadata represents object-level metadata
type ObjectMetadata struct {
	ID                     string          `json:"id,omitempty"`
	AppID                  *string         `json:"app_id,omitempty"`
	APIName                string          `json:"api_name"`
	Label                  string          `json:"label"`
	PluralLabel            string          `json:"plural_label"`
	Icon                   string          `json:"icon,omitempty"`
	Description            *string         `json:"description,omitempty"`
	IsSystem               bool            `json:"is_system,omitempty"`
	IsCustom               bool            `json:"is_custom,omitempty"`
	ThemeColor             *string         `json:"theme_color,omitempty"`
	SharingModel           SharingModel    `json:"sharing_model"`
	EnableHierarchySharing bool            `json:"enable_hierarchy_sharing"`
	Fields                 []FieldMetadata `json:"fields"`
	DefaultListView        *string         `json:"default_list_view,omitempty"`
	KanbanGroupBy          *string         `json:"kanban_group_by,omitempty"`
	KanbanSummaryField     *string         `json:"kanban_summary_field,omitempty"`
	ListFields             []string        `json:"list_fields,omitempty"`
	Searchable             bool            `json:"searchable"`
	PathField              *string         `json:"path_field,omitempty"` // Field to use for Path component (must be Picklist)
}

// ListView represents a list view configuration
type ListView struct {
	ID            string `json:"id"`
	ObjectAPIName string `json:"object_api_name"`
	Label         string `json:"label"`

	FilterExpr string   `json:"filter_expr,omitempty"`
	Fields     []string `json:"fields,omitempty"`
}

// PageLayout represents page layout configuration
type PageLayout struct {
	ID               string              `json:"id"`
	ObjectAPIName    string              `json:"object_api_name"`
	LayoutName       string              `json:"layout_name"`
	Type             string              `json:"type,omitempty"` // Detail, Edit, Create, List
	IsDefault        bool                `json:"is_default,omitempty"`
	CompactLayout    []string            `json:"compact_layout"`
	Tabs             []string            `json:"tabs,omitempty"`
	Sections         []PageSection       `json:"sections"`
	RelatedLists     []RelatedListConfig `json:"related_lists"`
	HeaderActions    []ActionConfig      `json:"header_actions"`
	QuickActions     []ActionConfig      `json:"quick_actions"`
	CreatedDate      time.Time           `json:"created_date,omitempty"`
	LastModifiedDate time.Time           `json:"last_modified_date,omitempty"`
}

// PageSection represents a section in a page layout
type PageSection struct {
	ID                  string                 `json:"id"`
	Label               string                 `json:"label"`
	Type                *string                `json:"type,omitempty"` // Fields, Component
	ComponentName       *string                `json:"component_name,omitempty"`
	ComponentConfig     map[string]interface{} `json:"component_config,omitempty"`
	Columns             int                    `json:"columns"`
	Fields              []string               `json:"fields"`
	VisibilityCondition *string                `json:"visibility_condition,omitempty"`
}

// RelatedListConfig represents a related list configuration
type RelatedListConfig struct {
	ID            string   `json:"id"`
	Label         string   `json:"label"`
	ObjectAPIName string   `json:"object_api_name"`
	LookupField   string   `json:"lookup_field"`
	Fields        []string `json:"fields"`
}

// ActionConfig represents an action configuration
type ActionConfig struct {
	Name                string                 `json:"name"`
	Label               string                 `json:"label"`
	Type                string                 `json:"type"` // Standard, CreateRecord, UpdateRecord, Flow, Custom, Url
	Icon                *string                `json:"icon,omitempty"`
	TargetObject        *string                `json:"target_object,omitempty"`
	Config              map[string]interface{} `json:"config,omitempty"`
	VisibilityCondition *string                `json:"visibility_condition,omitempty"`
	Component           *string                `json:"component,omitempty"`
}

// ActionMetadata represents action metadata
type ActionMetadata struct {
	ID            string                 `json:"id"`
	ObjectAPIName string                 `json:"object_api_name"`
	Name          string                 `json:"name"`
	Label         string                 `json:"label"`
	Type          string                 `json:"type"`
	Icon          string                 `json:"icon"`
	TargetObject  *string                `json:"target_object,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
}

type RecordType struct {
	ID                string    `json:"id"`
	ObjectAPIName     string    `json:"object_api_name"`
	Name              string    `json:"name"`
	Label             string    `json:"label"`
	Description       *string   `json:"description,omitempty"`
	IsActive          bool      `json:"is_active"`
	IsDefault         bool      `json:"is_default"`
	BusinessProcessID *string   `json:"business_process_id,omitempty"`
	CreatedDate       time.Time `json:"created_date"`
	LastModifiedDate  time.Time `json:"last_modified_date"`
}

type AutoNumber struct {
	ID               string    `json:"id"`
	ObjectAPIName    string    `json:"object_api_name"`
	FieldAPIName     string    `json:"field_api_name"`
	DisplayFormat    string    `json:"display_format"`
	StartingNumber   int       `json:"starting_number"`
	CurrentValue     int       `json:"current_value"`
	CreatedDate      time.Time `json:"created_date"`
	LastModifiedDate time.Time `json:"last_modified_date"`
}

type Relationship struct {
	ID                  string    `json:"id"`
	ChildObjectAPIName  string    `json:"child_object_api_name"`
	ParentObjectAPIName string    `json:"parent_object_api_name"`
	FieldAPIName        string    `json:"field_api_name"`
	RelationshipName    string    `json:"relationship_name"`
	RelationshipType    string    `json:"relationship_type"`
	CascadeDelete       bool      `json:"cascade_delete"`
	RestrictedDelete    bool      `json:"restricted_delete"`
	RelatedListLabel    *string   `json:"related_list_label,omitempty"`
	RelatedListFields   *string   `json:"related_list_fields,omitempty"`
	CreatedDate         time.Time `json:"created_date"`
	LastModifiedDate    time.Time `json:"last_modified_date"`
}

type FieldDependency struct {
	ID               string    `json:"id"`
	ObjectAPIName    string    `json:"object_api_name"`
	ControllingField string    `json:"controlling_field"`
	DependentField   string    `json:"dependent_field"`
	ControllingValue string    `json:"controlling_value"`
	Action           string    `json:"action"`
	IsActive         bool      `json:"is_active"`
	CreatedDate      time.Time `json:"created_date"`
	LastModifiedDate time.Time `json:"last_modified_date"`
}

// ValidationRule represents a validation rule
type ValidationRule struct {
	ID            string `json:"id"`
	ObjectAPIName string `json:"object_api_name"`
	Name          string `json:"name"`
	Active        bool   `json:"active"`
	Condition     string `json:"condition"`
	ErrorMessage  string `json:"error_message"`
}

// NavigationItem represents a navigation item in an app
type NavigationItem struct {
	ID            string `json:"id"`
	Type          string `json:"type"` // object, page, web
	ObjectAPIName string `json:"object_api_name,omitempty"`
	PageURL       string `json:"page_url,omitempty"`
	DashboardID   string `json:"dashboard_id,omitempty"`
	Label         string `json:"label"`
	Icon          string `json:"icon"`
}

// AppConfig represents application configuration
type AppConfig struct {
	ID               string           `json:"id"`
	Label            string           `json:"label"`
	Description      string           `json:"description"`
	Icon             string           `json:"icon"`
	Color            string           `json:"color"`
	IsDefault        bool             `json:"is_default"`
	NavigationItems  []NavigationItem `json:"navigation_items,omitempty"`
	CreatedDate      time.Time        `json:"created_date,omitempty"`
	LastModifiedDate time.Time        `json:"last_modified_date,omitempty"`
}

// Theme represents a visual theme
type Theme struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	IsActive         bool                   `json:"is_active"`
	Colors           map[string]interface{} `json:"colors"`
	Density          string                 `json:"density"`
	LogoURL          *string                `json:"logo_url,omitempty"`
	CreatedDate      time.Time              `json:"created_date,omitempty"`
	LastModifiedDate time.Time              `json:"last_modified_date,omitempty"`
}

// DashboardConfig represents dashboard configuration
type DashboardConfig struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Description *string        `json:"description,omitempty"`
	Layout      string         `json:"layout,omitempty"`
	Widgets     []WidgetConfig `json:"widgets"`
}

// WidgetConfig represents a dashboard widget
type WidgetConfig struct {
	ID     string                 `json:"id"`
	Title  string                 `json:"title"`
	Type   string                 `json:"type"`
	Query  AnalyticsQuery         `json:"query"`
	Config map[string]interface{} `json:"config,omitempty"` // Flexible config for widget-specific settings (e.g., SQL queries)
	X      *int                   `json:"x,omitempty"`
	Y      *int                   `json:"y,omitempty"`
	W      *int                   `json:"w,omitempty"`
	H      *int                   `json:"h,omitempty"`
	Icon   *string                `json:"icon,omitempty"`
	Color  *string                `json:"color,omitempty"`
}

type ProfileLayoutAssignment struct {
	LayoutID string `json:"layout_id"`
}

// SetupPage represents a page in the setup area
type SetupPage struct {
	ID                 string    `json:"id"`
	Label              string    `json:"label"`
	Icon               string    `json:"icon"`
	ComponentName      string    `json:"component_name"`
	Category           string    `json:"category"`
	PageOrder          int       `json:"page_order"`
	PermissionRequired string    `json:"permission_required,omitempty"`
	IsEnabled          bool      `json:"is_enabled"`
	Description        string    `json:"description,omitempty"`
	CreatedDate        time.Time `json:"created_date"`
	LastModifiedDate   time.Time `json:"last_modified_date"`
}

// UIComponent represents a registered UI component
type UIComponent struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"` // page, widget, etc.
	IsEmbeddable  bool      `json:"is_embeddable"`
	Description   *string   `json:"description,omitempty"`
	ComponentPath *string   `json:"component_path,omitempty"`
	CreatedDate   time.Time `json:"created_date,omitempty"`
	LastModified  time.Time `json:"last_modified_date,omitempty"`
}
