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
	ToolListDashboards   = "list_dashboards"
	ToolGetDashboard     = "get_dashboard"
	ToolCalculateFormula = "calculate_formula"
	ToolListThemes       = "list_themes"
	ToolActivateTheme    = "activate_theme"
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
		Description: "Query records from a specific object using a filter formula. Use 'filter' to specify conditions like \"status = 'Open'\" or \"amount > 1000\". Multiple conditions can be separated by AND.",
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
		Description: "Create a new record in any object/table. Use describe_object first to see required fields.",
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
		Description: "Update an existing record. Use search_records first to find the record ID.",
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
		Description: "Delete a record. Use search_records first to find the record ID. This action cannot be undone.",
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
				"widgets": map[string]interface{}{
					"type":        "array",
					"description": "Array of widget configurations. Each widget should have: title (string), type ('list', 'chart', 'metric', 'sql_chart'). For 'list' type: object, filter, columns. For 'chart' type: object, chart_type ('pie', 'bar', 'line'), group_by, agg_function ('count', 'sum', 'avg'). For 'metric' type: object, agg_field, agg_function. For 'sql_chart': sql query string.",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"title":        map[string]interface{}{"type": "string", "description": "Widget title"},
							"type":         map[string]interface{}{"type": "string", "description": "'list', 'chart', 'metric', or 'sql_chart'"},
							"object":       map[string]interface{}{"type": "string", "description": "Target object API name"},
							"filter":       map[string]interface{}{"type": "string", "description": "Filter expression"},
							"columns":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Columns"},
							"chart_type":   map[string]interface{}{"type": "string", "description": "'pie', 'bar', 'line'"},
							"group_by":     map[string]interface{}{"type": "string", "description": "Field to group by"},
							"agg_field":    map[string]interface{}{"type": "string", "description": "Field to aggregate"},
							"agg_function": map[string]interface{}{"type": "string", "description": "'count', 'sum', 'avg', 'min', 'max'"},
							"sql":          map[string]interface{}{"type": "string", "description": "SQL query for sql_chart type"},
							"size":         map[string]interface{}{"type": "string", "description": "'small', 'medium', 'large'"},
						},
						"required": []string{"title", "type"},
					},
				},
			},
			"required": []string{"name", "widgets"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolCreateObject,
		Description: "Create a new custom object/table. Example: Create a 'Vehicle' object.",
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
					"type": "boolean",
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
					"description": "Formula expression for Formula fields (e.g. \"amount * 0.1\")",
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
					"description": "The analytics operation (count, sum, avg, group_by)",
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
					"description": "Optional formula filter (e.g. \"status = 'Closed'\")",
				},
			},
			"required": []string{"object_api_name", "operation"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolListApps,
		Description: "List all application configurations, including their navigation items.",
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
					"type": "string",
				},
				"field_name": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"object_name", "field_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolGetRecord,
		Description: "Retrieve a specific record by its ID. Use this when you have a record ID and need its full data.",
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
					"type": "string",
				},
				"label": map[string]interface{}{
					"type": "string",
				},
				"description": map[string]interface{}{
					"type": "string",
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
					"type": "string",
				},
				"field_name": map[string]interface{}{
					"type": "string",
				},
				"label": map[string]interface{}{
					"type": "string",
				},
				"options": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "For Picklist types",
				},
			},
			"required": []string{"object_name", "field_name"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolUpdateApp,
		Description: "Update an application configuration (label, icon, description, navigation).",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
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
					"items": map[string]interface{}{
						"type": "object",
					},
				},
			},
			"required": []string{"id"},
		},
	})

	allTools = append(allTools, mcp.Tool{
		Name:        ToolDeleteApp,
		Description: "Delete an application configuration.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
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

	// Parse required fields
	name, ok := req.Arguments["name"].(string)
	if !ok || name == "" {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "name is required"}}}, nil
	}

	widgetsRaw, ok := req.Arguments["widgets"].([]interface{})
	if !ok {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "widgets array is required"}}}, nil
	}

	// Parse optional fields
	description, _ := req.Arguments["description"].(string)
	label, _ := req.Arguments["label"].(string)
	if label == "" {
		label = name
	}
	layout, _ := req.Arguments["layout"].(string)
	if layout == "" {
		layout = "two-column"
	}

	// Convert widgets to proper struct
	var widgets []models.DashboardWidget
	for _, w := range widgetsRaw {
		widgetMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		widget := models.DashboardWidget{
			Title:  getStringFromMap(widgetMap, "title"),
			Type:   getStringFromMap(widgetMap, "type"),
			Config: make(map[string]interface{}),
		}

		query := models.AnalyticsQuery{}

		if v := getStringFromMap(widgetMap, "object"); v != "" {
			query.ObjectAPIName = v
		}
		if v := getStringFromMap(widgetMap, "filter"); v != "" {
			query.FilterExpr = v
		}
		if cols, ok := widgetMap["columns"].([]interface{}); ok {
			var colStrings []string
			for _, c := range cols {
				if cs, ok := c.(string); ok {
					colStrings = append(colStrings, cs)
				}
			}
			widget.Config["columns"] = colStrings
		}
		if v := getStringFromMap(widgetMap, "chart_type"); v != "" {
			widget.Config["chart_type"] = v
		}
		if v := getStringFromMap(widgetMap, "group_by"); v != "" {
			val := v
			query.GroupBy = &val
		}
		if v := getStringFromMap(widgetMap, "agg_field"); v != "" {
			val := v
			query.Field = &val
		}
		if v := getStringFromMap(widgetMap, "agg_function"); v != "" {
			query.Operation = v
		}
		if v := getStringFromMap(widgetMap, "sql"); v != "" {
			widget.Config["sql"] = v
		}
		if v := getStringFromMap(widgetMap, "size"); v != "" {
			widget.Config["size"] = v
		}

		widget.Query = query
		widgets = append(widgets, widget)
	}

	dashboard := models.DashboardCreate{
		Label:       label,
		Description: &description,
		Layout:      layout,
		Widgets:     widgets,
	}

	id, err := s.client.CreateDashboard(ctx, dashboard, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Create dashboard failed: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created dashboard '%s' with ID: %s", name, id)}},
	}, nil
}

func (s *ToolBusService) handleCreateObject(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	apiName, ok1 := req.Arguments["api_name"].(string)
	label, ok2 := req.Arguments["label"].(string)
	pluralLabel, ok3 := req.Arguments["plural_label"].(string)

	if !ok1 || !ok2 || !ok3 {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "api_name, label, and plural_label are required"}}}, nil
	}

	description, _ := req.Arguments["description"].(string)

	schema := models.ObjectMetadata{
		APIName:     apiName,
		Label:       label,
		PluralLabel: pluralLabel,
		Description: &description,
		IsCustom:    true,
	}

	if err := s.client.CreateObject(ctx, schema, token); err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create object: %v", err)}}}, nil
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Successfully created object '%s' (%s)", label, apiName)}},
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

func (s *ToolBusService) handleCreateApp(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	name, ok1 := req.Arguments["name"].(string)
	label, ok2 := req.Arguments["label"].(string)

	if !ok1 || !ok2 {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "name and label are required"}}}, nil
	}

	description, _ := req.Arguments["description"].(string)

	var navItems []models.NavigationItem
	if items, ok := req.Arguments["nav_items"].([]interface{}); ok {
		for _, item := range items {
			if str, ok := item.(string); ok {
				navItems = append(navItems, models.NavigationItem{
					Type:          "object",
					ObjectAPIName: str,
					Label:         str, // Default label to API name
				})
			}
		}
	}

	app := models.AppConfig{
		Label:           label,
		Description:     description,
		NavigationItems: navItems,
	}

	id, err := s.client.CreateApp(ctx, app, token)
	if err != nil {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Failed to create app: %v", err)}}}, nil
	}

	msg := fmt.Sprintf("Successfully created app '%s' (%s)", label, name)
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

	objectName, _ := arguments["object_name"].(string)
	var schema models.ObjectMetadata
	if label, ok := arguments["label"].(string); ok {
		schema.Label = label
	}
	if desc, ok := arguments["description"].(string); ok {
		schema.Description = &desc
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
