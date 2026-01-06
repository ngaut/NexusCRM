package server

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/nexuscrm/mcp/pkg/client"
	"github.com/nexuscrm/mcp/pkg/contextstore"
	"github.com/nexuscrm/mcp/pkg/mcp"
	"github.com/nexuscrm/mcp/pkg/models"
)

const (
	ToolListObjects     = "list_objects"
	ToolDescribeObject  = "describe_object"
	ToolQueryObject     = "query_object"
	ToolCreateRecord    = "create_record"
	ToolUpdateRecord    = "update_record"
	ToolDeleteRecord    = "delete_record"
	ToolCreateDashboard = "create_dashboard"
	// Schema Tools
	ToolCreateObject = "create_object"
	ToolCreateField  = "create_field"
	ToolCreateApp    = "create_app"
	// Context Tools
	ToolContextAdd    = "context_add"
	ToolContextRemove = "context_remove"
	ToolContextList   = "context_list"
	ToolContextClear  = "context_clear"
	// Search & Analytics
	ToolSearchRecords = "search_records"
	ToolSearchObject  = "search_object_records"
	ToolRunAnalytics  = "run_analytics"
	ToolListApps      = "list_apps"
	// Deletion Tools
	ToolDeleteObject = "delete_object"
	ToolDeleteField  = "delete_field"
	// Record Retrieval
	ToolGetRecord = "get_record"
	// Update Tools
	ToolUpdateObject    = "update_object"
	ToolUpdateField     = "update_field"
	ToolUpdateApp       = "update_app"
	ToolUpdateDashboard = "update_dashboard"
	// Recycle Bin Tools
	ToolGetRecycleBin = "get_recycle_bin"
	ToolRestoreRecord = "restore_record"
	ToolPurgeRecord   = "purge_record"
	// Management
	ToolDeleteApp       = "delete_app"
	ToolDeleteDashboard = "delete_dashboard"
	// New Management Tools
	ToolListDashboards     = "list_dashboards"
	ToolGetDashboard       = "get_dashboard"
	ToolCalculateFormula   = "calculate_formula"
	ToolListThemes         = "list_themes"
	ToolActivateTheme      = "activate_theme"
	ToolAddDashboardWidget = "add_dashboard_widget"
	// Scheduled Jobs Tools
	ToolListScheduledJobs  = "list_scheduled_jobs"
	ToolCreateScheduledJob = "create_scheduled_job"
	ToolUpdateSchedule     = "update_schedule"
	ToolRunJobNow          = "run_job_now"
	// Validation Rule Tools
	ToolCreateValidationRule = "create_validation_rule"
	ToolUpdateValidationRule = "update_validation_rule"
	ToolDeleteValidationRule = "delete_validation_rule"
	ToolGetValidationRules   = "get_validation_rules"
)

type ToolBusService struct {
	client       *client.NexusClient
	contextStore *contextstore.ContextStore
}

func NewToolBusService(client *client.NexusClient, contextStore *contextstore.ContextStore) *ToolBusService {
	return &ToolBusService{
		client:       client,
		contextStore: contextStore,
	}
}

func (s *ToolBusService) getAuthToken(ctx context.Context) (string, error) {
	token, ok := ctx.Value(mcp.ContextKeyAuthToken).(string)
	if !ok || token == "" {
		return "", &mcp.Error{Code: mcp.ErrInternal, Message: "Unauthorized: No auth token"}
	}
	return token, nil
}

// HandleListTools returns discovery tools + generic CRUD tools
func (s *ToolBusService) HandleListTools(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var allTools []mcp.Tool

	// 1. Discovery Tools
	allTools = append(allTools, mcp.Tool{
		Name:        ToolListObjects,
		Description: "List all available objects/tables in the CRM. Use this FIRST to discover what data is available before searching or creating records.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Regex pattern to filter objects (case-insensitive). Matches against Name or Label.",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Max results to return (default 50)",
				},
			},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDescribeObject,
		Description: "Get the full schema for an object, including all fields and their types. Use this to understand what fields are required before creating or updating records.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object from list_objects (e.g., 'Account', 'jira_issue', '_System_User')",
				},
			},
			"required": []string{"object_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolQueryObject,
		Description: "Query business data records from a specific object. For dashboards use list_dashboards, for apps use list_apps instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object to search (e.g., 'Account', 'jira_issue')",
				},
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "Filter expression using formula syntax. Operators: ==, !=, >, <, >=, <=, &&, ||. String matching: CONTAINS(field, 'text'), STARTS_WITH(field, 'text'). Null checks: field == null (IS NULL), field != null (IS NOT NULL). Examples: \"status == 'Open'\", \"amount > 1000 && type == 'Enterprise'\". TIP: If query returns 0 but object exists, try use limit 1 without filter first to verify data exists.",
				},
				"sort_field": map[string]interface{}{
					"type":        "string",
					"description": "Field to sort by (e.g. 'created_date')",
				},
				"sort_order": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"ASC", "DESC"},
					"description": "Sort direction (default DESC)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Max results (default 20)",
				},
			},
			"required": []string{"object_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateRecord,
		Description: "Create a new business data record (e.g., Account, Contact, Lead). Use describe_object first to see required fields. DO NOT use for system objects - use dedicated tools (create_dashboard, create_app, create_object, create_field) instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "Field values for the new record",
				},
			},
			"required": []string{"object_name", "data"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateRecord,
		Description: "Update an existing business data record (e.g., Account, Contact, Lead). DO NOT use for system objects - use dedicated tools (update_dashboard, update_app, update_object, update_field) instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object",
				},
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The record ID to update",
				},
				"data": map[string]interface{}{
					"type":        "object",
					"description": "Fields to update",
				},
			},
			"required": []string{"object_name", "id", "data"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteRecord,
		Description: "Delete a business data record (e.g., Account, Contact, Lead). Moves to recycle bin. DO NOT use for system objects like _System_Dashboard, _System_App, etc. - use their dedicated tools (delete_dashboard, delete_app) instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object",
				},
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The record ID to delete",
				},
			},
			"required": []string{"object_name", "id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateDashboard,
		Description: "Create a dashboard with widgets. Use this specialized tool instead of create_record for _System_Dashboard. Widgets are passed as a structured array, NOT as a JSON string.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard name (required)",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard label (optional, defaults to name)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard description",
				},
				"layout": map[string]interface{}{
					"type":        "string",
					"description": "Layout type: 'two-column', 'grid', or 'single'",
					"default":     "two-column",
				},
			},

			"required": []string{"name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolAddDashboardWidget,
		Description: "Add a widget to an existing dashboard.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dashboard_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the dashboard to modify",
				},
				"widget": map[string]interface{}{
					"type":        "object",
					"description": "Widget configuration",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{"type": "string", "description": "Widget title"},
						"type":  map[string]interface{}{"type": "string", "enum": []string{"metric", "chart-bar", "chart-pie", "chart-line", "chart-funnel", "record-list", "kanban", "sql-chart"}, "description": "Widget type"},
						"query": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"object_api_name": map[string]interface{}{"type": "string", "description": "Target object API name (e.g., 'opportunity', 'lead')"},
								"operation":       map[string]interface{}{"type": "string", "enum": []string{"count", "sum", "avg", "min", "max", "group_by"}, "description": "Aggregation operation"},
								"field":           map[string]interface{}{"type": "string", "description": "Field to aggregate (for sum/avg)"},
								"group_by":        map[string]interface{}{"type": "string", "description": "Group by field (for charts)"},
								"filter_expr":     map[string]interface{}{"type": "string", "description": "Optional filter using expr-lang syntax (e.g., \"status == 'Open'\")"},
							},
						},
						"config": map[string]interface{}{
							"type":        "object",
							"description": "Widget-specific config (e.g., chart_type, columns, sql, content, imageUrl)",
						},
						"x":     map[string]interface{}{"type": "integer", "description": "Grid X position (0-11)"},
						"y":     map[string]interface{}{"type": "integer", "description": "Grid Y position"},
						"w":     map[string]interface{}{"type": "integer", "description": "Grid Width (1-12)"},
						"h":     map[string]interface{}{"type": "integer", "description": "Grid Height"},
						"icon":  map[string]interface{}{"type": "string", "description": "Icon name (e.g. 'Users')"},
						"color": map[string]interface{}{"type": "string", "description": "Widget accent color (hex or name)"},
					},
					"required": []string{"title", "type"},
				},
			},
			"required": []string{"dashboard_id", "widget"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateObject,
		Description: "Create a new custom object/table. Example: Create a 'Vehicle' object. NOTE: After creating an object, you may want to use 'update_app' to add it to the navigation menu so users can see it.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"api_name": map[string]interface{}{
					"type":        "string",
					"description": "API name (snake_case, e.g. 'vehicle'). Must be unique.",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Human readable label (e.g. 'Vehicle')",
				},
				"plural_label": map[string]interface{}{
					"type":        "string",
					"description": "Plural label (e.g. 'Vehicles')",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the object",
				},
				"sharing_model": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"Private", "PublicRead", "PublicReadWrite"},
					"description": "Object sharing model (default Private)",
				},
				"icon": map[string]interface{}{
					"type":        "string",
					"description": "Lucide icon name (e.g. 'Box', 'User')",
				},
				"theme_color": map[string]interface{}{
					"type":        "string",
					"description": "Theme color hex code or name (e.g. '#FF0000', 'blue')",
				},
				"enable_hierarchy_sharing": map[string]interface{}{
					"type":        "boolean",
					"description": "Use hierarchy for sharing access",
				},
			},
			"required": []string{"api_name", "label", "plural_label"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateField,
		Description: "Create a new field on an existing object.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object (e.g. 'account')",
				},
				"api_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the field (snake_case, e.g. 'model_year')",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Field label",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"Text", "Number", "Currency", "Boolean", "Date", "DateTime", "Email", "Phone", "URL", "Picklist", "Lookup", "RollupSummary", "Formula", "TextArea", "LongTextArea", "RichText", "Percent", "JSON", "Password", "AutoNumber"},
					"description": "Field type. Use 'Formula' for calculated fields.",
				},
				"required": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether the field is required (default false)",
				},
				"options": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Options for Picklist type",
				},
				"reference_to": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Target object for Lookup type (e.g. ['account'])",
				},
				"formula_expression": map[string]interface{}{
					"type":        "string",
					"description": "Formula expression using expr-lang syntax (https://expr-lang.org/). Use field API names directly. Examples: \"amount * 0.1\", \"first_name + ' ' + last_name\", \"amount > 10000 ? 'High' : 'Low'\"",
				},
				"default_value": map[string]interface{}{
					"type":        "string",
					"description": "Default value for the field",
				},
			},
			"required": []string{"object_name", "api_name", "label", "type"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolContextAdd,
		Description: "Add files to the conversation context. The content of these files will be available to the AI.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"files": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of file paths to add",
				},
			},
			"required": []string{"files"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolContextRemove,
		Description: "Remove files from the conversation context.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"files": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of file paths to remove",
				},
			},
			"required": []string{"files"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolContextList,
		Description: "List all files currently in the conversation context.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolContextClear,
		Description: "Clear all files from the conversation context.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolSearchRecords,
		Description: "Perform a global text search across all searchable objects in the CRM. Use this for broad queries like finding a person's name or a company across different tables.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"term": map[string]interface{}{
					"type":        "string",
					"description": "The search term (e.g. 'John Doe', 'Acme Corp')",
				},
			},
			"required": []string{"term"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolSearchObject,
		Description: "Perform a text search within a specific object. Use this when you know which object to search but want to find records matching a text string.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object to search",
				},
				"term": map[string]interface{}{
					"type":        "string",
					"description": "The search term",
				},
			},
			"required": []string{"object_name", "term"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolRunAnalytics,
		Description: "Run an analytics query to get counts, sums, or group-by results. Use this for reports and metrics.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_api_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object to analyze",
				},
				"operation": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"count", "sum", "avg", "min", "max", "group_by"},
					"description": "Aggregation operation",
				},
				"field": map[string]interface{}{
					"type":        "string",
					"description": "The field to aggregate (for sum/avg)",
				},
				"group_by": map[string]interface{}{
					"type":        "string",
					"description": "The field to group by (for group_by)",
				},
				"filter_expr": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter using expr-lang syntax (https://expr-lang.org/). Examples: \"status == 'Closed'\", \"amount > 10000\"",
				},
			},
			"required": []string{"object_api_name", "operation"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolListApps,
		Description: "List all application configurations, including their navigation items. Call this FIRST when you need to add items to an app's navigation/sidebar - you need the app ID for update_app.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteObject,
		Description: "Delete an object schema. WARNING: This will also delete all data in the object.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "The API name of the object to delete",
				},
			},
			"required": []string{"object_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteField,
		Description: "Permanently delete a custom field from an object schema.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object (e.g., 'account')",
				},
				"field_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the field to delete",
				},
			},
			"required": []string{"object_name", "field_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolGetRecord,
		Description: "Retrieve a specific business data record by its ID. For dashboards use get_dashboard, for apps use the list_apps tool instead.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object (e.g., 'Account')",
				},
				"id": map[string]interface{}{
					"type":        "string",
					"description": "The unique UUID of the record",
				},
			},
			"required": []string{"object_name", "id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateObject,
		Description: "Update properties of an existing object schema (e.g., label, description).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object to update",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "New display label",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "New description",
				},
				"plural_label": map[string]interface{}{
					"type":        "string",
					"description": "New plural label",
				},
				"icon": map[string]interface{}{
					"type":        "string",
					"description": "New icon name",
				},
				"theme_color": map[string]interface{}{
					"type":        "string",
					"description": "New theme color",
				},
				"sharing_model": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"Private", "PublicRead", "PublicReadWrite"},
					"description": "New sharing model",
				},
				"enable_hierarchy_sharing": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable/Disable hierarchy sharing",
				},
			},
			"required": []string{"object_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateField,
		Description: "Update properties of an existing field schema (e.g., label, options).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object containing the field",
				},
				"field_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the field to update",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "New display label",
				},
				"options": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "New picklist options (for Picklist type fields)",
				},
			},
			"required": []string{"object_name", "field_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateApp,
		Description: "Update an application configuration. Use this to add objects, dashboards, or web links to the app's navigation/sidebar/menu (also called 'navigator items'). WORKFLOW: 1) First call list_apps to get the app ID, 2) Then call update_app with navigation_items array. NOTE: navigation_items REPLACES all existing items, so include both old and new items.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "App ID (e.g. 'nexus_crm')",
				},
				"label": map[string]interface{}{
					"type": "string",
				},
				"icon": map[string]interface{}{
					"type": "string",
				},
				"description": map[string]interface{}{
					"type": "string",
				},
				"navigation_items": map[string]interface{}{
					"type": "array",
					"description": `List of navigation items. Each item has a 'type' that determines required fields:
- type='object': Links to an object list view. Requires 'object_api_name' (e.g. 'account', 'contact').
- type='dashboard': Links to a dashboard. Requires 'dashboard_id'.
- type='web': External web link. Requires 'page_url' (e.g. 'https://example.com' or '/dashboard').
All items require 'type' and 'label'. Examples:
  {"type": "object", "label": "Accounts", "object_api_name": "account"}
  {"type": "dashboard", "label": "Sales Metrics", "dashboard_id": "dash-123"}
  {"type": "web", "label": "Help Center", "page_url": "https://help.example.com"}`,
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"type": map[string]interface{}{
								"type":        "string",
								"enum":        []string{"object", "dashboard", "web"},
								"description": "Type of navigation item: 'object' for data tables, 'dashboard' for analytics, 'web' for external links",
							},
							"label": map[string]interface{}{
								"type":        "string",
								"description": "Display label in the navigation menu",
							},
							"object_api_name": map[string]interface{}{
								"type":        "string",
								"description": "Required if type='object'. The API name of the object (e.g. 'account', 'contact', 'opportunity')",
							},
							"dashboard_id": map[string]interface{}{
								"type":        "string",
								"description": "Required if type='dashboard'. The ID of the dashboard to link to",
							},
							"page_url": map[string]interface{}{
								"type":        "string",
								"description": "Required if type='web'. URL for external link (e.g. 'https://help.example.com')",
							},
							"icon": map[string]interface{}{
								"type":        "string",
								"description": "Lucide icon name (e.g. 'Database', 'LayoutDashboard', 'Globe')",
							},
						},
						"required": []string{"type", "label"},
					},
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateApp,
		Description: "Create a new application with navigation menu. Navigation can include objects (data tables), dashboards, and web links.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Unique app ID (snake_case, e.g. 'sales_app')",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Display name (e.g. 'Sales App')",
				},
				"description": map[string]interface{}{
					"type": "string",
				},
				"icon": map[string]interface{}{
					"type":        "string",
					"description": "Lucide icon name (e.g. 'Layers', 'Briefcase')",
				},
				"navigation_items": map[string]interface{}{
					"type": "array",
					"description": `List of navigation items. Can be simple strings (object API names) or full objects.
Simple: ["account", "contact"] - creates object links with auto-labels.
Full objects for more control:
  {"type": "object", "label": "Accounts", "object_api_name": "account"}
  {"type": "dashboard", "label": "Metrics", "dashboard_id": "dash-123"}
  {"type": "web", "label": "Docs", "page_url": "https://docs.example.com"}`,
					"items": map[string]interface{}{
						"anyOf": []interface{}{
							map[string]interface{}{
								"type":        "string",
								"description": "Object API name (shorthand for type='object')",
							},
							map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"type":            map[string]interface{}{"type": "string", "enum": []string{"object", "dashboard", "web"}, "description": "'object'=data table, 'dashboard'=analytics, 'web'=external link"},
									"label":           map[string]interface{}{"type": "string", "description": "Menu display label"},
									"object_api_name": map[string]interface{}{"type": "string", "description": "For type='object': API name (e.g. 'account')"},
									"dashboard_id":    map[string]interface{}{"type": "string", "description": "For type='dashboard': Dashboard ID"},
									"page_url":        map[string]interface{}{"type": "string", "description": "For type='web': External URL"},
									"icon":            map[string]interface{}{"type": "string", "description": "Lucide icon name"},
								},
								"required": []string{"type", "label"},
							},
						},
					},
				},
			},
			"required": []string{"id", "label"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteApp,
		Description: "Delete an application configuration.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "App ID to delete (e.g., 'sales_app')",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolGetRecycleBin,
		Description: "List items in the recycle bin for inspection or restoration.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"mine", "all"},
					"description": "Default is 'mine'",
				},
			},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolRestoreRecord,
		Description: "Restore a record from the recycle bin back to its original object.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Recycle bin item ID (not the original record ID)",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolPurgeRecord,
		Description: "Permanently delete a record from the recycle bin. This cannot be undone.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Recycle bin item ID",
				},
			},
			"required": []string{"id"},
		},
	})

	// Dashboard Management Tools
	allTools = append(allTools, mcp.Tool{
		Name:        ToolListDashboards,
		Description: "List all dashboards in the system. Use this to find dashboard IDs for get_dashboard, update_dashboard, or delete_dashboard.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolGetDashboard,
		Description: "Get a specific dashboard by ID including its widgets configuration.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard ID",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateDashboard,
		Description: "Update an existing dashboard's name, description, layout, or widgets.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard ID to update",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "New dashboard name/label",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "New description",
				},
				"layout": map[string]interface{}{
					"type":        "string",
					"description": "Layout: 'two-column', 'grid', or 'single'",
				},
				"widgets": map[string]interface{}{
					"type":        "array",
					"description": "Updated widgets array",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteDashboard,
		Description: "Delete a dashboard. Use list_dashboards to find dashboard IDs first.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Dashboard ID to delete",
				},
			},
			"required": []string{"id"},
		},
	})

	// Formula & Theme Tools
	allTools = append(allTools, mcp.Tool{
		Name: ToolCalculateFormula,
		Description: `Evaluate a formula expression with optional record context. Uses expr-lang syntax (https://expr-lang.org/).

SYNTAX: Use field API names directly for record context. Supports arithmetic (+, -, *, /), comparisons (==, !=, >, <), logical operators (&&, ||, !), ternary (? :), and string concatenation (+).

EXAMPLES:
- "amount * 0.1" - Calculate 10% of amount
- "first_name + ' ' + last_name" - Concatenate names
- "probability > 50 ? 'Likely' : 'Unlikely'" - Conditional`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "expr-lang formula expression to evaluate",
				},
				"object_name": map[string]interface{}{
					"type":        "string",
					"description": "Optional: object context for field references",
				},
				"record_id": map[string]interface{}{
					"type":        "string",
					"description": "Optional: record ID for field value substitution",
				},
			},
			"required": []string{"expression"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolListThemes,
		Description: "List all available UI themes.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolActivateTheme,
		Description: "Activate a UI theme by ID. Only one theme can be active at a time.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "Theme ID to activate",
				},
			},
			"required": []string{"id"},
		},
	})

	// Scheduled Jobs Tools
	allTools = append(allTools, mcp.Tool{
		Name:        ToolListScheduledJobs,
		Description: "List all scheduled jobs (flows with trigger_type='schedule'). Returns job name, schedule (cron expression), last run time, next run time, and status.",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateScheduledJob,
		Description: "Create a new scheduled job. This creates a flow that runs on a cron schedule. Common schedules: '0 9 * * *' (daily at 9AM), '*/15 * * * *' (every 15 mins), '0 0 1 * *' (monthly).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the scheduled job",
				},
				"schedule": map[string]interface{}{
					"type":        "string",
					"description": "Cron expression (5 fields: minute hour day-of-month month day-of-week)",
				},
				"action_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of action to execute: 'update_field', 'call_webhook', 'send_email'",
				},
				"action_config": map[string]interface{}{
					"type":        "object",
					"description": "Configuration for the action (depends on action_type)",
				},
				"timezone": map[string]interface{}{
					"type":        "string",
					"description": "Optional timezone for schedule (default: UTC). Example: 'America/New_York'",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional description of what this job does",
				},
			},
			"required": []string{"name", "schedule", "action_type", "action_config"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateSchedule,
		Description: "Update the schedule of an existing scheduled job. Can also enable/disable the job.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the scheduled job to update",
				},
				"schedule": map[string]interface{}{
					"type":        "string",
					"description": "New cron expression (optional)",
				},
				"timezone": map[string]interface{}{
					"type":        "string",
					"description": "New timezone (optional)",
				},
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Set to 'Active' to enable or 'Inactive' to disable",
				},
			},
			"required": []string{"job_id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolRunJobNow,
		Description: "Manually trigger a scheduled job to run immediately, regardless of its schedule.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"job_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the scheduled job to run",
				},
			},
			"required": []string{"job_id"},
		},
	})

	// Validation Rule Tools
	allTools = append(allTools, mcp.Tool{
		Name: ToolCreateValidationRule,
		Description: `Create a validation rule for an object. The condition uses expr-lang syntax (https://expr-lang.org/) and evaluates to TRUE when the error should be shown.

SYNTAX GUIDE:
- Field access: Use field API names directly (e.g., amount, close_date, stage)
- Null checks: field == null or field != null
- Comparisons: ==, !=, >, <, >=, <=
- Logical: && (and), || (or), ! (not)
- Strings: Use single quotes 'text'
- Empty strings for Date/Number fields are auto-converted to null

EXAMPLES:
- "amount < 0" - Amount cannot be negative
- "stage == 'Closed' && close_date == null" - Close date required for closed deals
- "probability > 100 || probability < 0" - Probability must be 0-100`,
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_api_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object (e.g. 'opportunity')",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Unique name for the rule (e.g. 'Close_Date_Required')",
				},
				"error_message": map[string]interface{}{
					"type":        "string",
					"description": "The error message to display to the user",
				},
				"condition": map[string]interface{}{
					"type":        "string",
					"description": "expr-lang formula expression. Returns TRUE to trigger the error. Use field names directly, 'null' for null checks. Example: \"stage == 'Closed' && close_date == null\"",
				},
				"active": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether the rule is active (default true)",
				},
			},
			"required": []string{"object_api_name", "name", "error_message", "condition"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateValidationRule,
		Description: "Update an existing validation rule.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the validation rule",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "New name (optional)",
				},
				"error_message": map[string]interface{}{
					"type":        "string",
					"description": "New error message (optional)",
				},
				"condition": map[string]interface{}{
					"type":        "string",
					"description": "New formula condition (optional)",
				},
				"active": map[string]interface{}{
					"type":        "boolean",
					"description": "New active status (optional)",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteValidationRule,
		Description: "Delete a validation rule.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the validation rule to delete",
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolGetValidationRules,
		Description: "List validation rules for a specific object.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"object_api_name": map[string]interface{}{
					"type":        "string",
					"description": "API name of the object",
				},
			},
			"required": []string{"object_api_name"},
		},
	})

	return mcp.ListToolsResult{Tools: allTools}, nil
}

// HandleCallTool executes a tool
func (s *ToolBusService) HandleCallTool(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req mcp.CallToolParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params"}
	}

	// Tool routing based on tool name
	switch req.Name {
	case ToolListObjects:
		return s.handleListObjects(ctx, req)
	case ToolDescribeObject:
		return s.handleDescribeObject(ctx, req)
	case ToolQueryObject:
		return s.handleQueryObject(ctx, req)
	case ToolCreateRecord:
		return s.handleCreateRecord(ctx, req)
	case ToolUpdateRecord:
		return s.handleUpdateRecord(ctx, req)
	case ToolDeleteRecord:
		return s.handleDeleteRecord(ctx, req)
	case ToolCreateDashboard:
		return s.handleCreateDashboard(ctx, req)
	case ToolAddDashboardWidget:
		return s.handleAddDashboardWidget(ctx, req)
	case ToolSearchRecords:
		return s.handleSearchRecords(ctx, req)
	case ToolSearchObject:
		return s.handleSearchObject(ctx, req)
	case ToolRunAnalytics:
		return s.handleRunAnalytics(ctx, req)
	case ToolListApps:
		return s.handleListApps(ctx, req)
	case ToolCreateObject:
		return s.handleCreateObject(ctx, req)
	case ToolDeleteObject:
		return s.handleDeleteObject(ctx, req)
	case ToolCreateField:
		return s.handleCreateField(ctx, req)
	case ToolDeleteField:
		return s.handleDeleteField(ctx, req.Arguments)
	case ToolCreateApp:
		return s.handleCreateApp(ctx, req)
	case ToolContextAdd:
		return s.handleContextAdd(ctx, req)
	case ToolContextRemove:
		return s.handleContextRemove(ctx, req)
	case ToolContextList:
		return s.handleContextList(ctx, req)
	case ToolContextClear:
		return s.handleContextClear(ctx, req)
	case ToolGetRecord:
		return s.handleGetRecord(ctx, req.Arguments)
	case ToolUpdateObject:
		return s.handleUpdateObject(ctx, req.Arguments)
	case ToolUpdateField:
		return s.handleUpdateField(ctx, req.Arguments)
	case ToolUpdateApp:
		return s.handleUpdateApp(ctx, req.Arguments)
	case ToolDeleteApp:
		return s.handleDeleteApp(ctx, req.Arguments)
	case ToolUpdateDashboard:
		return s.handleUpdateDashboard(ctx, req.Arguments)
	case ToolDeleteDashboard:
		return s.handleDeleteDashboard(ctx, req.Arguments)
	case ToolGetRecycleBin:
		return s.handleGetRecycleBin(ctx, req.Arguments)
	case ToolRestoreRecord:
		return s.handleRestoreRecord(ctx, req.Arguments)
	case ToolPurgeRecord:
		return s.handlePurgeRecord(ctx, req.Arguments)
	case ToolListDashboards:
		return s.handleListDashboards(ctx, req.Arguments)
	case ToolGetDashboard:
		return s.handleGetDashboard(ctx, req.Arguments)
	case ToolCalculateFormula:
		return s.handleCalculateFormula(ctx, req.Arguments)
	case ToolListThemes:
		return s.handleListThemes(ctx, req.Arguments)
	case ToolActivateTheme:
		return s.handleActivateTheme(ctx, req.Arguments)
	case ToolListScheduledJobs:
		return s.handleListScheduledJobs(ctx, req.Arguments)
	case ToolCreateScheduledJob:
		return s.handleCreateScheduledJob(ctx, req.Arguments)
	case ToolUpdateSchedule:
		return s.handleUpdateSchedule(ctx, req.Arguments)
	case ToolRunJobNow:
		return s.handleRunJobNow(ctx, req.Arguments)
	case ToolCreateValidationRule:
		return s.handleCreateValidationRule(ctx, req.Arguments)
	case ToolUpdateValidationRule:
		return s.handleUpdateValidationRule(ctx, req.Arguments)
	case ToolDeleteValidationRule:
		return s.handleDeleteValidationRule(ctx, req.Arguments)
	case ToolGetValidationRules:
		return s.handleGetValidationRules(ctx, req.Arguments)
	default:
		return nil, &mcp.Error{Code: mcp.ErrMethodNotFound, Message: fmt.Sprintf("Tool '%s' not found", req.Name)}
	}
}

// handleListObjects returns a list of objects via API
func (s *ToolBusService) handleListObjects(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	// Parse parameters
	query := ""
	limit := 50
	if val, ok := req.Arguments["query"].(string); ok {
		query = val
	}
	if val, ok := req.Arguments["limit"].(float64); ok {
		limit = int(val)
	}

	objects, err := s.client.ListObjects(ctx, token)
	if err != nil {
		return mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to list objects: %v", err)}},
			IsError: true,
		}, nil
	}

	type ObjectSummary struct {
		Name        string `json:"name"`
		Label       string `json:"label"`
		Description string `json:"description,omitempty"`
	}

	var filtered []ObjectSummary
	regexPattern := "(?i)" + query
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		regex = regexp.MustCompile("(?i)" + regexp.QuoteMeta(query))
	}

	count := 0
	for _, obj := range objects {
		if query != "" {
			if !regex.MatchString(obj.APIName) && !regex.MatchString(obj.Label) {
				continue
			}
		}

		desc := ""
		if obj.Description != nil {
			desc = *obj.Description
		}
		filtered = append(filtered, ObjectSummary{
			Name:        obj.APIName,
			Label:       obj.Label,
			Description: desc,
		})

		count++
		if count >= limit {
			break
		}
	}

	jsonBytes, _ := json.MarshalIndent(filtered, "", "  ")

	msg := fmt.Sprintf("Found %d objects", len(filtered))
	if len(objects) > count {
		msg += fmt.Sprintf(" (showing top %d)", count)
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("%s:\n%s\n\nUse describe_object to get full field details.", msg, string(jsonBytes))}},
	}, nil
}

func (s *ToolBusService) handleDescribeObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := req.Arguments["object_name"].(string)
	if !ok || objectName == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name required"}}}, nil
	}

	meta, err := s.client.DescribeObject(ctx, objectName, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Describe failed: %v", err)}}}, nil
	}

	jsonBytes, _ := json.MarshalIndent(meta, "", "  ")
	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}},
	}, nil
}

func (s *ToolBusService) handleQueryObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := req.Arguments["object_name"].(string)
	if !ok || objectName == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name required"}}}, nil
	}

	// Use filter expression directly - let the formula engine handle parsing
	filterExpr, _ := req.Arguments["filter"].(string)

	limit := 20
	if l, ok := req.Arguments["limit"].(float64); ok {
		limit = int(l)
	}

	sortField, _ := req.Arguments["sort_field"].(string)
	sortOrder, _ := req.Arguments["sort_order"].(string)

	queryReq := models.QueryRequest{
		ObjectAPIName: objectName,
		FilterExpr:    filterExpr,
		Limit:         limit,
		SortField:     sortField,
		SortDirection: sortOrder,
	}

	results, err := s.client.Query(ctx, queryReq, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Query failed: %v", err)}}}, nil
	}

	if len(results) == 0 {
		return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("No records found for %s", objectName)}}}, nil
	}

	jsonBytes, _ := json.MarshalIndent(results, "", "  ")
	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Found %d records:\n%s", len(results), string(jsonBytes))}},
	}, nil
}

func (s *ToolBusService) handleCreateRecord(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := req.Arguments["object_name"].(string)
	data, okData := req.Arguments["data"].(map[string]interface{})
	if !ok || !okData {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name and data required"}}}, nil
	}

	id, err := s.client.CreateRecord(ctx, objectName, data, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Create failed: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created %s record with ID: %s", objectName, id)}},
	}, nil
}

func (s *ToolBusService) handleUpdateRecord(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := req.Arguments["object_name"].(string)
	id, okId := req.Arguments["id"].(string)
	data, okData := req.Arguments["data"].(map[string]interface{})

	if !ok || !okId || !okData {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name, id, and data required"}}}, nil
	}

	err = s.client.UpdateRecord(ctx, objectName, id, data, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Update failed: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully updated %s record %s", objectName, id)}},
	}, nil
}

func (s *ToolBusService) handleDeleteRecord(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := req.Arguments["object_name"].(string)
	id, okId := req.Arguments["id"].(string)

	if !ok || !okId {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name and id required"}}}, nil
	}

	err = s.client.DeleteRecord(ctx, objectName, id, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Delete failed: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully deleted %s record %s", objectName, id)}},
	}, nil
}

func (s *ToolBusService) handleCreateDashboard(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	// Root Fix: Use JSON serialization to robustly map all fields (including layout, config, content)
	// instead of manual extraction which causes data loss.
	jsonBytes, err := json.Marshal(req.Arguments)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to parse params: %v", err)}}}, nil
	}

	var startDashboard models.DashboardCreate
	if err := json.Unmarshal(jsonBytes, &startDashboard); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid dashboard parameters: %v", err)}}}, nil
	}

	id, err := s.client.CreateDashboard(ctx, startDashboard, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create dashboard: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created dashboard '%s' (ID: %s)", startDashboard.Label, id)}},
	}, nil
}

func (s *ToolBusService) handleAddDashboardWidget(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	dashboardID, ok := req.Arguments["dashboard_id"].(string)
	if !ok || dashboardID == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "dashboard_id is required"}}}, nil
	}

	widgetMap, ok := req.Arguments["widget"].(map[string]interface{})
	if !ok {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "widget configuration is required"}}}, nil
	}

	// Unmarshal widget
	widgetBytes, err := json.Marshal(widgetMap)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to marshal widget: %v", err)}}}, nil
	}
	var widget models.DashboardWidget
	if err := json.Unmarshal(widgetBytes, &widget); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid widget format: %v", err)}}}, nil
	}

	// Fetch Dashboard
	dashboard, err := s.client.GetDashboard(ctx, dashboardID, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to fetch dashboard: %v", err)}}}, nil
	}

	// Append
	dashboard.Widgets = append(dashboard.Widgets, widget)

	// Update
	if err := s.client.UpdateDashboard(ctx, dashboardID, *dashboard, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to update dashboard: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully added widget '%s' to dashboard", widget.Title)}},
	}, nil
}

func (s *ToolBusService) handleCreateObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	// Root Fix: Use JSON serialization
	jsonBytes, err := json.Marshal(req.Arguments)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to parse params: %v", err)}}}, nil
	}

	var schema models.ObjectMetadata
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid object parameters: %v", err)}}}, nil
	}

	// Ensure IsCustom is defaults to true unless specified (backend usually handles logic, but nice to be explicit)
	schema.IsCustom = true

	if err := s.client.CreateObject(ctx, schema, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create object: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created object '%s' (%s)", schema.Label, schema.APIName)}},
	}, nil
}

func (s *ToolBusService) handleCreateField(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok1 := req.Arguments["object_name"].(string)
	apiName, ok2 := req.Arguments["api_name"].(string)
	label, ok3 := req.Arguments["label"].(string)
	fieldType, ok4 := req.Arguments["type"].(string)

	if !ok1 || !ok2 || !ok3 || !ok4 {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name, api_name, label, and type are required"}}}, nil
	}

	field := models.FieldMetadata{
		APIName: apiName,
		Label:   label,
		Type:    models.FieldType(fieldType), // Potentially risky cast, but backed by API validation
	}

	if val, ok := req.Arguments["required"].(bool); ok {
		field.Required = val
	}

	if opts, ok := req.Arguments["options"].([]interface{}); ok {
		for _, o := range opts {
			if str, ok := o.(string); ok {
				field.Options = append(field.Options, str)
			}
		}
	}

	if refs, ok := req.Arguments["reference_to"].([]interface{}); ok {
		for _, r := range refs {
			if str, ok := r.(string); ok {
				field.ReferenceTo = append(field.ReferenceTo, str)
			}
		}
	}

	if err := s.client.CreateField(ctx, objectName, field, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create field: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created field '%s' on %s", apiName, objectName)}},
	}, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getStringFromSObject extracts a string field from models.SObject
func getStringFromSObject(m models.SObject, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (s *ToolBusService) handleCreateApp(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	appID, ok1 := req.Arguments["id"].(string)
	label, ok2 := req.Arguments["label"].(string)

	if !ok1 || appID == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "id is required (e.g., 'sales_app')"}}}, nil
	}
	if !ok2 || label == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "label is required (e.g., 'Sales App')"}}}, nil
	}

	description, _ := req.Arguments["description"].(string)
	icon, _ := req.Arguments["icon"].(string)

	var navItems []models.NavigationItem
	itemsRaw := req.Arguments["navigation_items"]

	if items, ok := itemsRaw.([]interface{}); ok {
		for _, item := range items {
			// Handle string input (API Name alias)
			if str, ok := item.(string); ok {
				navItems = append(navItems, models.NavigationItem{
					Type:          "object",
					ObjectAPIName: str,
					Label:         str, // Default label to API name (will be humanized by UI if needed)
				})
				continue
			}

			// Handle object input (Full definition)
			if mapVal, ok := item.(map[string]interface{}); ok {
				var navItem models.NavigationItem
				// We need to carefully convert map to struct
				// Using json marshal/unmarshal is the safest lazy way
				jsonBytes, _ := json.Marshal(mapVal)
				if err := json.Unmarshal(jsonBytes, &navItem); err == nil {
					// Validate required fields based on type
					if navItem.Type == "" {
						navItem.Type = "object" // Default
					}
					// Ensure ID is generated if missing (backend might handle this, but let's be safe)
					// Actually backend/models structure probably uses DB ID.
					// We'll let backend handle ID generation.
					navItems = append(navItems, navItem)
				}
			}
		}
	}

	app := models.AppConfig{
		ID:              appID,
		Label:           label,
		Description:     description,
		Icon:            icon,
		NavigationItems: navItems,
	}

	id, err := s.client.CreateApp(ctx, app, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create app: %v", err)}}}, nil
	}

	msg := fmt.Sprintf("Successfully created app '%s' (%s)", label, appID)
	if id != "" {
		msg += fmt.Sprintf(" with ID: %s", id)
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: msg}},
	}, nil
}

func (s *ToolBusService) handleSearchRecords(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	term, _ := req.Arguments["term"].(string)
	if term == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "term is required"}}}, nil
	}
	results, err := s.client.Search(ctx, term, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Search failed: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(results, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleSearchObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	obj, _ := req.Arguments["object_name"].(string)
	term, _ := req.Arguments["term"].(string)
	if obj == "" || term == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name and term are required"}}}, nil
	}
	results, err := s.client.SearchObject(ctx, obj, term, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Search failed: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(results, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleRunAnalytics(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	obj, _ := req.Arguments["object_api_name"].(string)
	op, _ := req.Arguments["operation"].(string)
	if obj == "" || op == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_api_name and operation are required"}}}, nil
	}
	query := models.AnalyticsQuery{
		ObjectAPIName: obj,
		Operation:     op,
		FilterExpr:    getStringFromMap(req.Arguments, "filter_expr"),
	}
	if f, ok := req.Arguments["field"].(string); ok {
		query.Field = &f
	}
	if g, ok := req.Arguments["group_by"].(string); ok {
		query.GroupBy = &g
	}

	result, err := s.client.RunAnalytics(ctx, query, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Analytics failed: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleListApps(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	apps, err := s.client.ListApps(ctx, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("List apps failed: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(apps, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleDeleteObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	obj, _ := req.Arguments["object_name"].(string)
	if obj == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name is required"}}}, nil
	}
	if err := s.client.DeleteObject(ctx, obj, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Delete failed: %v", err)}}}, nil
	}
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully deleted object %s", obj)}}}, nil
}

func (s *ToolBusService) handleDeleteField(ctx context.Context, apiArgs map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	obj, _ := apiArgs["object_name"].(string)
	field, _ := apiArgs["field_name"].(string)
	if obj == "" || field == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name and field_name are required"}}}, nil
	}
	if err := s.client.DeleteField(ctx, obj, field, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Delete failed: %v", err)}}}, nil
	}
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully deleted field %s on %s", field, obj)}}}, nil
}

func (s *ToolBusService) handleGetRecord(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, _ := arguments["object_name"].(string)
	id, _ := arguments["id"].(string)

	record, err := s.client.GetRecord(ctx, objectName, id, token)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error retrieving record: %v", err)}},
		}, nil
	}

	jsonData, _ := json.MarshalIndent(record, "", "  ")
	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: string(jsonData)}},
	}, nil
}

func (s *ToolBusService) handleUpdateObject(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, ok := arguments["object_name"].(string)
	if !ok || objectName == "" {
		// Fallback check if 'api_name' was used (common AI mistake due to create_object usage)
		if n, ok := arguments["api_name"].(string); ok {
			objectName = n
		} else {
			return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_name is required"}}}, nil
		}
	}

	// Root Fix: Use JSON serialization
	// Note: We strip object_name/api_name from map? No need, extra fields ignored by Unmarshal or overwrite APIName field which is fine. (we use objectName var for URL)
	jsonBytes, err := json.Marshal(arguments)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to parse params: %v", err)}}}, nil
	}

	var schema models.ObjectMetadata
	if err := json.Unmarshal(jsonBytes, &schema); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid update parameters: %v", err)}}}, nil
	}

	if err := s.client.UpdateObject(ctx, objectName, schema, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error updating object: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Object updated successfully"}},
	}, nil
}

func (s *ToolBusService) handleUpdateField(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectName, _ := arguments["object_name"].(string)
	fieldName, _ := arguments["field_name"].(string)

	var field models.FieldMetadata
	if label, ok := arguments["label"].(string); ok {
		field.Label = label
	}
	if optionsRaw, ok := arguments["options"].([]interface{}); ok {
		options := make([]string, len(optionsRaw))
		for i, v := range optionsRaw {
			options[i], _ = v.(string)
		}
		field.Options = options
	}

	if err := s.client.UpdateField(ctx, objectName, fieldName, field, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error updating field: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Field updated successfully"}},
	}, nil
}

func (s *ToolBusService) handleUpdateApp(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	var app models.AppConfig
	if label, ok := arguments["label"].(string); ok {
		app.Label = label
	}
	if icon, ok := arguments["icon"].(string); ok {
		app.Icon = icon
	}
	if desc, ok := arguments["description"].(string); ok {
		app.Description = desc
	}
	if navRaw, ok := arguments["navigation_items"].([]interface{}); ok {
		navData, _ := json.Marshal(navRaw)
		var navItems []models.NavigationItem
		json.Unmarshal(navData, &navItems)
		app.NavigationItems = navItems
	}

	if err := s.client.UpdateApp(ctx, id, app, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error updating app: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "App updated successfully"}},
	}, nil
}

func (s *ToolBusService) handleDeleteApp(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	if err := s.client.DeleteApp(ctx, id, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error deleting app: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "App deleted successfully"}},
	}, nil
}

func (s *ToolBusService) handleUpdateDashboard(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	var dashboard models.DashboardConfig
	data, _ := json.Marshal(arguments)
	json.Unmarshal(data, &dashboard)

	if err := s.client.UpdateDashboard(ctx, id, dashboard, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error updating dashboard: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Dashboard updated successfully"}},
	}, nil
}

func (s *ToolBusService) handleDeleteDashboard(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	if err := s.client.DeleteDashboard(ctx, id, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error deleting dashboard: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Dashboard deleted successfully"}},
	}, nil
}

func (s *ToolBusService) handleGetRecycleBin(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	scope, _ := arguments["scope"].(string)
	if scope == "" {
		scope = "mine"
	}

	items, err := s.client.GetRecycleBinItems(ctx, scope, token)
	if err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error retrieving recycle bin: %v", err)}},
		}, nil
	}

	jsonData, _ := json.MarshalIndent(items, "", "  ")
	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: string(jsonData)}},
	}, nil
}

func (s *ToolBusService) handleRestoreRecord(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	if err := s.client.RestoreRecord(ctx, id, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error restoring record: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Record restored successfully"}},
	}, nil
}

func (s *ToolBusService) handlePurgeRecord(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := arguments["id"].(string)
	if err := s.client.PurgeRecord(ctx, id, token); err != nil {
		return mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error purging record: %v", err)}},
		}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Record purged successfully"}},
	}, nil
}

func (s *ToolBusService) handleListDashboards(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	dashboards, err := s.client.GetDashboards(ctx, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error listing dashboards: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(dashboards, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleGetDashboard(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	id, _ := arguments["id"].(string)
	dashboard, err := s.client.GetDashboard(ctx, id, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error getting dashboard: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(dashboard, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleCalculateFormula(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	obj, _ := arguments["object_name"].(string)
	record, _ := arguments["record"].(map[string]interface{})
	result, err := s.client.Calculate(ctx, obj, record, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error calculating formula: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleListThemes(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	theme, err := s.client.ListThemes(ctx, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error listing themes: %v", err)}}}, nil
	}
	jsonBytes, _ := json.MarshalIndent(theme, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

func (s *ToolBusService) handleActivateTheme(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}
	id, _ := arguments["id"].(string)
	if err := s.client.ActivateTheme(ctx, id, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error activating theme: %v", err)}}}, nil
	}
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: "Theme activated successfully"}}}, nil
}

// handleListScheduledJobs returns all scheduled jobs (flows with trigger_type='schedule')
func (s *ToolBusService) handleListScheduledJobs(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	// Query _system_flow with trigger_type = 'schedule'
	req := models.QueryRequest{
		ObjectAPIName: "_system_flow",
		FilterExpr:    "trigger_type = 'schedule'",
		Limit:         100,
	}
	records, err := s.client.Query(ctx, req, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error querying scheduled jobs: %v", err)}}}, nil
	}

	// Format response for better readability
	type ScheduledJob struct {
		ID               string  `json:"id"`
		Name             string  `json:"name"`
		Schedule         string  `json:"schedule"`
		ScheduleTimezone string  `json:"schedule_timezone"`
		LastRunAt        *string `json:"last_run_at,omitempty"`
		NextRunAt        *string `json:"next_run_at,omitempty"`
		Status           string  `json:"status"`
		IsRunning        bool    `json:"is_running"`
		ActionType       string  `json:"action_type"`
	}

	var jobs []ScheduledJob
	for _, rec := range records {
		job := ScheduledJob{
			ID:         getStringFromSObject(rec, "id"),
			Name:       getStringFromSObject(rec, "name"),
			Schedule:   getStringFromSObject(rec, "schedule"),
			Status:     getStringFromSObject(rec, "status"),
			ActionType: getStringFromSObject(rec, "action_type"),
		}
		if tz := getStringFromSObject(rec, "schedule_timezone"); tz != "" {
			job.ScheduleTimezone = tz
		}
		if lastRun := getStringFromSObject(rec, "last_run_at"); lastRun != "" {
			job.LastRunAt = &lastRun
		}
		if nextRun := getStringFromSObject(rec, "next_run_at"); nextRun != "" {
			job.NextRunAt = &nextRun
		}
		if running, ok := rec["is_running"].(bool); ok {
			job.IsRunning = running
		}
		jobs = append(jobs, job)
	}

	result := map[string]interface{}{
		"total_jobs": len(jobs),
		"jobs":       jobs,
	}
	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

// handleCreateScheduledJob creates a new scheduled flow
func (s *ToolBusService) handleCreateScheduledJob(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	// Extract required parameters
	name, _ := arguments["name"].(string)
	schedule, _ := arguments["schedule"].(string)
	actionType, _ := arguments["action_type"].(string)
	actionConfig, _ := arguments["action_config"].(map[string]interface{})

	if name == "" || schedule == "" || actionType == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "Missing required parameters: name, schedule, action_type"}}}, nil
	}

	// Validate cron expression format (basic check)
	cronPattern := regexp.MustCompile(`^(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)$`)
	if !cronPattern.MatchString(schedule) {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid cron expression: '%s'. Expected format: 'minute hour day-of-month month day-of-week'", schedule)}}}, nil
	}

	// Build flow record
	flowData := map[string]interface{}{
		"name":         name,
		"trigger_type": "schedule",
		"schedule":     schedule,
		"action_type":  actionType,
		"flow_type":    "simple",
		"status":       "Active",
	}

	if actionConfig != nil {
		flowData["action_config"] = actionConfig
	}

	// Optional timezone
	if timezone, ok := arguments["timezone"].(string); ok && timezone != "" {
		flowData["schedule_timezone"] = timezone
	}

	// Optional description
	if description, ok := arguments["description"].(string); ok && description != "" {
		flowData["description"] = description
	}

	// Optional trigger_object (some actions need it)
	flowData["trigger_object"] = "_system_config" // Default object for scheduled jobs

	// Create the flow
	jobID, err := s.client.CreateRecord(ctx, "_system_flow", flowData, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error creating scheduled job: %v", err)}}}, nil
	}

	jsonBytes, _ := json.MarshalIndent(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Scheduled job '%s' created successfully", name),
		"job_id":  jobID,
	}, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

// handleUpdateSchedule updates an existing scheduled job
func (s *ToolBusService) handleUpdateSchedule(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	jobID, _ := arguments["job_id"].(string)
	if jobID == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "Missing required parameter: job_id"}}}, nil
	}

	// Build update data
	updateData := make(map[string]interface{})

	if schedule, ok := arguments["schedule"].(string); ok && schedule != "" {
		// Validate cron expression
		cronPattern := regexp.MustCompile(`^(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)\s+(\*|[0-9,\-\/]+)$`)
		if !cronPattern.MatchString(schedule) {
			return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Invalid cron expression: '%s'", schedule)}}}, nil
		}
		updateData["schedule"] = schedule
		// Clear next_run_at so scheduler recalculates
		updateData["next_run_at"] = nil
	}

	if timezone, ok := arguments["timezone"].(string); ok && timezone != "" {
		updateData["schedule_timezone"] = timezone
	}

	if status, ok := arguments["status"].(string); ok && status != "" {
		if status != "Active" && status != "Inactive" {
			return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "Status must be 'Active' or 'Inactive'"}}}, nil
		}
		updateData["status"] = status
	}

	if len(updateData) == 0 {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "No updates provided. Specify schedule, timezone, or status."}}}, nil
	}

	// Update the flow
	err = s.client.UpdateRecord(ctx, "_system_flow", jobID, updateData, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error updating scheduled job: %v", err)}}}, nil
	}

	jsonBytes, _ := json.MarshalIndent(map[string]interface{}{
		"success": true,
		"message": "Scheduled job updated successfully",
		"job_id":  jobID,
	}, "", "  ")
	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}}}, nil
}

// handleRunJobNow triggers a scheduled job to run immediately
func (s *ToolBusService) handleRunJobNow(ctx context.Context, arguments map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	jobID, _ := arguments["job_id"].(string)
	if jobID == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "Missing required parameter: job_id"}}}, nil
	}

	// Call the flow execution endpoint
	err = s.client.ExecuteFlow(ctx, jobID, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error running job: %v", err)}}}, nil
	}

	return mcp.CallToolResult{Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully triggered job %s", jobID)}}}, nil
}

// handleCreateValidationRule handles creation of a new validation rule
func (s *ToolBusService) handleCreateValidationRule(ctx context.Context, args map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectAPIName, _ := args["object_api_name"].(string)
	name, _ := args["name"].(string)
	errorMsg, _ := args["error_message"].(string)
	condition, _ := args["condition"].(string)
	active, _ := args["active"].(bool)

	// Default active to true if not specified? Schema says so
	if _, ok := args["active"]; !ok {
		active = true
	}

	if objectAPIName == "" || name == "" || errorMsg == "" || condition == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_api_name, name, error_message, and condition are required"}}}, nil
	}

	rule := models.ValidationRule{
		ObjectAPIName: objectAPIName,
		Name:          name,
		ErrorMessage:  errorMsg,
		Condition:     condition,
		Active:        active,
	}

	id, err := s.client.CreateValidationRule(ctx, rule, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create validation rule: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created validation rule '%s' with ID: %s", name, id)}},
	}, nil
}

// handleUpdateValidationRule handles updating an existing validation rule
func (s *ToolBusService) handleUpdateValidationRule(ctx context.Context, args map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := args["id"].(string)
	if id == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "id is required"}}}, nil
	}

	rule := models.ValidationRule{}
	if val, ok := args["name"].(string); ok {
		rule.Name = val
	}
	if val, ok := args["error_message"].(string); ok {
		rule.ErrorMessage = val
	}
	if val, ok := args["condition"].(string); ok {
		rule.Condition = val
	}
	if val, ok := args["active"].(bool); ok {
		rule.Active = val
	}

	if err := s.client.UpdateValidationRule(ctx, id, rule, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to update validation rule: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully updated validation rule %s", id)}},
	}, nil
}

// handleDeleteValidationRule handles deleting a validation rule
func (s *ToolBusService) handleDeleteValidationRule(ctx context.Context, args map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	id, _ := args["id"].(string)
	if id == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "id is required"}}}, nil
	}

	if err := s.client.DeleteValidationRule(ctx, id, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to delete validation rule: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully deleted validation rule %s", id)}},
	}, nil
}

// handleGetValidationRules returns validation rules for an object
func (s *ToolBusService) handleGetValidationRules(ctx context.Context, args map[string]interface{}) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	objectAPIName, _ := args["object_api_name"].(string)
	if objectAPIName == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "object_api_name is required"}}}, nil
	}

	rules, err := s.client.GetValidationRules(ctx, objectAPIName, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to get validation rules: %v", err)}}}, nil
	}

	jsonBytes, _ := json.MarshalIndent(rules, "", "  ")

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: string(jsonBytes)}},
	}, nil
}
