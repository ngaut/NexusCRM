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
	// Context Tools
	ToolContextAdd    = "context_add"
	ToolContextRemove = "context_remove"
	ToolContextList   = "context_list"
	ToolContextClear  = "context_clear"
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
					"enum":        []string{"Text", "Number", "Currency", "Boolean", "Date", "DateTime", "Email", "Phone", "URL", "Select", "Lookup", "Rollup"},
					"description": "Field type",
				},
				"required": map[string]interface{}{
					"type": "boolean",
				},
				"options": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Options for Select type",
				},
				"reference_to": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Target object for Lookup type (e.g. ['account'])",
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

	return mcp.ListToolsResult{Tools: allTools}, nil
}

// HandleCallTool executes a tool
func (s *ToolBusService) HandleCallTool(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req mcp.CallToolParams
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params"}
	}

	// Tool routing based on tool name
	if req.Name == ToolListObjects {
		return s.handleListObjects(ctx, req)
	}
	if req.Name == ToolDescribeObject {
		return s.handleDescribeObject(ctx, req)
	}
	if req.Name == ToolQueryObject {
		return s.handleQueryObject(ctx, req)
	}
	if req.Name == ToolCreateRecord {
		return s.handleCreateRecord(ctx, req)
	}
	if req.Name == ToolUpdateRecord {
		return s.handleUpdateRecord(ctx, req)
	}
	if req.Name == ToolDeleteRecord {
		return s.handleDeleteRecord(ctx, req)
	}
	if req.Name == ToolCreateDashboard {
		return s.handleCreateDashboard(ctx, req)
	}
	if req.Name == ToolCreateDashboard {
		return s.handleCreateDashboard(ctx, req)
	}
	// Schema Tools
	if req.Name == ToolCreateObject {
		return s.handleCreateObject(ctx, req)
	}
	if req.Name == ToolCreateField {
		return s.handleCreateField(ctx, req)
	}

	// Context Tools
	if req.Name == ToolContextAdd {
		return s.handleContextAdd(ctx, req)
	}
	if req.Name == ToolContextRemove {
		return s.handleContextRemove(ctx, req)
	}
	if req.Name == ToolContextList {
		return s.handleContextList(ctx, req)
	}
	if req.Name == ToolContextClear {
		return s.handleContextClear(ctx, req)
	}

	return nil, &mcp.Error{Code: mcp.ErrMethodNotFound, Message: fmt.Sprintf("Tool '%s' not found", req.Name)}
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
			Title: getStringFromMap(widgetMap, "title"),
			Type:  getStringFromMap(widgetMap, "type"),
		}
		if v := getStringFromMap(widgetMap, "object"); v != "" {
			widget.Object = v
		}
		if v := getStringFromMap(widgetMap, "filter"); v != "" {
			widget.Filter = v
		}
		if cols, ok := widgetMap["columns"].([]interface{}); ok {
			for _, c := range cols {
				if cs, ok := c.(string); ok {
					widget.Columns = append(widget.Columns, cs)
				}
			}
		}
		if v := getStringFromMap(widgetMap, "chart_type"); v != "" {
			widget.ChartType = v
		}
		if v := getStringFromMap(widgetMap, "group_by"); v != "" {
			widget.GroupBy = v
		}
		if v := getStringFromMap(widgetMap, "agg_field"); v != "" {
			widget.AggField = v
		}
		if v := getStringFromMap(widgetMap, "agg_function"); v != "" {
			widget.AggFunction = v
		}
		if v := getStringFromMap(widgetMap, "sql"); v != "" {
			widget.SQL = v
		}
		if v := getStringFromMap(widgetMap, "size"); v != "" {
			widget.Size = v
		}
		widgets = append(widgets, widget)
	}

	dashboard := models.DashboardCreate{
		Name:        name,
		Label:       label,
		Description: description,
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
