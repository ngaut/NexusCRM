package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nexuscrm/mcp/pkg/models"
)

type NexusClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewNexusClient(baseURL string) *NexusClient {
	return &NexusClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Helper to execute requests
func (c *NexusClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}, authToken string) error {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	fullURL := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		// Try both Cookie and Header for maximum compatibility
		req.Header.Set("Cookie", fmt.Sprintf("auth_token=%s", authToken))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBytes, _ := io.ReadAll(resp.Body)
		// Try to parse JSON error response for cleaner error messages
		var errResp struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBytes, &errResp) == nil {
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Error
			}
			if msg != "" {
				return fmt.Errorf("api error (%d): %s", resp.StatusCode, msg)
			}
		}
		// Fallback to raw response if JSON parsing fails or no message field
		if len(respBytes) > 0 {
			return fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBytes))
		}
		return fmt.Errorf("api error (%d): no details provided", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// API Methods

func (c *NexusClient) ListObjects(ctx context.Context, authToken string) ([]models.ObjectMetadata, error) {
	// Current API: GET /api/metadata/objects
	// Response envelope: { "schemas": [ ... ] }
	var respMap map[string][]models.ObjectMetadata

	if err := c.doRequest(ctx, "GET", "/api/metadata/objects", nil, &respMap, authToken); err != nil {
		return nil, err
	}

	if objs, ok := respMap["data"]; ok {
		return objs, nil
	}
	return nil, fmt.Errorf("invalid response format for list objects")
}

func (c *NexusClient) DescribeObject(ctx context.Context, objectName string, authToken string) (*models.ObjectMetadata, error) {
	// GET /api/metadata/objects/:name
	// Returns { "schema": ... }
	var respMap map[string]*models.ObjectMetadata
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/metadata/objects/%s", objectName), nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if obj, ok := respMap["data"]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("invalid response format for describe object")
}

func (c *NexusClient) Query(ctx context.Context, req models.QueryRequest, authToken string) ([]models.SObject, error) {
	// POST /api/data/query
	var respMap map[string][]models.SObject
	if err := c.doRequest(ctx, "POST", "/api/data/query", req, &respMap, authToken); err != nil {
		return nil, err
	}

	if records, ok := respMap["data"]; ok {
		return records, nil
	}
	return nil, fmt.Errorf("invalid response format for query")
}

func (c *NexusClient) CreateRecord(ctx context.Context, objectName string, data map[string]interface{}, authToken string) (string, error) {
	// POST /api/data/:objectApiName

	// Use interface{} to handle both { "data": object } and { "id": "..." } patterns
	var rawResp map[string]interface{}
	if err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/data/%s", objectName), data, &rawResp, authToken); err != nil {
		return "", err
	}

	// Check "data" -> "id"
	if dataVal, ok := rawResp["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if id, ok := dataMap["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("created record missing ID")
}

// GetRecord retrieves a single record by ID
func (c *NexusClient) GetRecord(ctx context.Context, objectName, id string, authToken string) (models.SObject, error) {
	// GET /api/data/:objectApiName/:id
	var respMap map[string]models.SObject
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/data/%s/%s", objectName, id), nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if record, ok := respMap["data"]; ok {
		return record, nil
	}
	return nil, fmt.Errorf("record not found")
}

func (c *NexusClient) UpdateRecord(ctx context.Context, objectName, id string, data map[string]interface{}, authToken string) error {
	// PATCH /api/data/:objectApiName/:id
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/data/%s/%s", objectName, id), data, nil, authToken)
}

func (c *NexusClient) DeleteRecord(ctx context.Context, objectName, id string, authToken string) error {
	// DELETE /api/data/:objectApiName/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/data/%s/%s", objectName, id), nil, nil, authToken)
}

// GetDashboards returns all dashboards visible to the user
func (c *NexusClient) GetDashboards(ctx context.Context, authToken string) ([]models.DashboardConfig, error) {
	// GET /api/metadata/dashboards
	var respMap map[string][]models.DashboardConfig
	if err := c.doRequest(ctx, "GET", "/api/metadata/dashboards", nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if dashboards, ok := respMap["data"]; ok {
		return dashboards, nil
	}
	return nil, fmt.Errorf("invalid response format for dashboards")
}

// GetDashboard retrieves a single dashboard by ID
func (c *NexusClient) GetDashboard(ctx context.Context, id string, authToken string) (*models.DashboardConfig, error) {
	// GET /api/metadata/dashboards/:id
	var respMap map[string]*models.DashboardConfig
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/api/metadata/dashboards/%s", id), nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if dashboard, ok := respMap["data"]; ok {
		return dashboard, nil
	}
	// Fallback check if it returns unwrapped? Handler returns { dashboard: ... }
	return nil, fmt.Errorf("dashboard not found")
}

func (c *NexusClient) CreateDashboard(ctx context.Context, dashboard models.DashboardCreate, authToken string) (string, error) {
	// POST /api/metadata/dashboards
	var rawResp map[string]interface{}
	if err := c.doRequest(ctx, "POST", "/api/metadata/dashboards", dashboard, &rawResp, authToken); err != nil {
		return "", err
	}

	// Check "data" -> "id" (Standard wrapper)
	if dataVal, ok := rawResp["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if id, ok := dataMap["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("created dashboard missing ID")
}

// CreateApp creates a new application configuration
func (c *NexusClient) CreateApp(ctx context.Context, app models.AppConfig, authToken string) (string, error) {
	// POST /api/metadata/apps
	var rawResp map[string]interface{}
	if err := c.doRequest(ctx, "POST", "/api/metadata/apps", app, &rawResp, authToken); err != nil {
		return "", err
	}

	// Check "data" -> "id"
	if dataVal, ok := rawResp["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if id, ok := dataMap["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("created app missing ID")
}

// UpdateApp updates an application configuration
func (c *NexusClient) UpdateApp(ctx context.Context, id string, app models.AppConfig, authToken string) error {
	// PATCH /api/metadata/apps/:id
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/metadata/apps/%s", id), app, nil, authToken)
}

// DeleteApp deletes an application configuration
func (c *NexusClient) DeleteApp(ctx context.Context, id string, authToken string) error {
	// DELETE /api/metadata/apps/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/metadata/apps/%s", id), nil, nil, authToken)
}

// CreateObject creates a new object schema
func (c *NexusClient) CreateObject(ctx context.Context, schema models.ObjectMetadata, authToken string) error {
	// POST /api/metadata/objects
	return c.doRequest(ctx, "POST", "/api/metadata/objects", schema, nil, authToken)
}

// DeleteObject deletes an object schema
func (c *NexusClient) DeleteObject(ctx context.Context, apiName string, authToken string) error {
	// DELETE /api/metadata/objects/:name
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/metadata/objects/%s", apiName), nil, nil, authToken)
}

// UpdateObject updates an object schema
func (c *NexusClient) UpdateObject(ctx context.Context, apiName string, schema models.ObjectMetadata, authToken string) error {
	// PATCH /api/metadata/objects/:apiName
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/metadata/objects/%s", apiName), schema, nil, authToken)
}

// CreateField creates a new field on an object
func (c *NexusClient) CreateField(ctx context.Context, objectName string, field models.FieldMetadata, authToken string) error {
	// POST /api/metadata/objects/:apiName/fields
	return c.doRequest(ctx, "POST", fmt.Sprintf("/api/metadata/objects/%s/fields", objectName), field, nil, authToken)
}

// DeleteField deletes a field from an object
func (c *NexusClient) DeleteField(ctx context.Context, objectName, fieldName string, authToken string) error {
	// DELETE /api/metadata/objects/:apiName/fields/:fieldName
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/metadata/objects/%s/fields/%s", objectName, fieldName), nil, nil, authToken)
}

// UpdateField updates a field on an object
func (c *NexusClient) UpdateField(ctx context.Context, objectName, fieldName string, field models.FieldMetadata, authToken string) error {
	// PATCH /api/metadata/objects/:apiName/fields/:fieldName
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/metadata/objects/%s/fields/%s", objectName, fieldName), field, nil, authToken)
}

// Search performs a global text search
func (c *NexusClient) Search(ctx context.Context, term string, authToken string) ([]interface{}, error) {
	// POST /api/data/search
	req := struct {
		Term string `json:"term"`
	}{Term: term}
	var respMap map[string][]interface{}
	if err := c.doRequest(ctx, "POST", "/api/data/search", req, &respMap, authToken); err != nil {
		return nil, err
	}
	if results, ok := respMap["data"]; ok {
		return results, nil
	}
	return nil, fmt.Errorf("invalid response format for search")
}

// Calculate previews a formula outcome for a record
func (c *NexusClient) Calculate(ctx context.Context, objectName string, data map[string]interface{}, authToken string) (map[string]interface{}, error) {
	// POST /api/data/:object/calculate
	var respMap map[string]map[string]interface{}
	if err := c.doRequest(ctx, "POST", fmt.Sprintf("/api/data/%s/calculate", objectName), data, &respMap, authToken); err != nil {
		return nil, err
	}
	if record, ok := respMap["data"]; ok {
		return record, nil
	}
	return nil, fmt.Errorf("invalid response format for calculate")
}

// GetValidationRules retrieves validation rules for an object
func (c *NexusClient) GetValidationRules(ctx context.Context, objectName string, authToken string) ([]interface{}, error) {
	// GET /api/metadata/validation-rules?objectApiName=...
	var respMap map[string][]interface{}
	path := fmt.Sprintf("/api/metadata/validation-rules?objectApiName=%s", objectName)
	if err := c.doRequest(ctx, "GET", path, nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if rules, ok := respMap["data"]; ok {
		return rules, nil
	}
	return nil, fmt.Errorf("invalid response format for validation rules")
}

// ListThemes returns the active theme
func (c *NexusClient) ListThemes(ctx context.Context, authToken string) (interface{}, error) {
	// GET /api/metadata/theme
	var respMap map[string]interface{}
	if err := c.doRequest(ctx, "GET", "/api/metadata/theme", nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if theme, ok := respMap["data"]; ok {
		return theme, nil
	}
	return nil, fmt.Errorf("invalid response format for theme")
}

// ActivateTheme switches the UI theme
func (c *NexusClient) ActivateTheme(ctx context.Context, id string, authToken string) error {
	// PUT /api/metadata/themes/:id/activate
	return c.doRequest(ctx, "PUT", fmt.Sprintf("/api/metadata/themes/%s/activate", id), nil, nil, authToken)
}

// SearchObject performs a text search within a specific object
func (c *NexusClient) SearchObject(ctx context.Context, objectName, term string, authToken string) ([]models.SObject, error) {
	// GET /api/data/search/:object?term=...
	var respMap map[string][]models.SObject
	path := fmt.Sprintf("/api/data/search/%s?term=%s", objectName, term)
	if err := c.doRequest(ctx, "GET", path, nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if records, ok := respMap["data"]; ok {
		return records, nil
	}
	return nil, fmt.Errorf("invalid response format for object search")
}

// RunAnalytics executes an analytics query
func (c *NexusClient) RunAnalytics(ctx context.Context, query models.AnalyticsQuery, authToken string) (interface{}, error) {
	// POST /api/data/analytics
	var respMap map[string]interface{}
	if err := c.doRequest(ctx, "POST", "/api/data/analytics", query, &respMap, authToken); err != nil {
		return nil, err
	}
	if result, ok := respMap["data"]; ok {
		return result, nil
	}
	return nil, fmt.Errorf("invalid response format for analytics")
}

// ListApps returns all application configurations
func (c *NexusClient) ListApps(ctx context.Context, authToken string) ([]models.AppConfig, error) {
	// GET /api/metadata/apps
	var respMap map[string][]models.AppConfig
	if err := c.doRequest(ctx, "GET", "/api/metadata/apps", nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if apps, ok := respMap["data"]; ok {
		return apps, nil
	}
	return nil, fmt.Errorf("invalid response format for list apps")
}

// UpdateDashboard updates a dashboard configuration
func (c *NexusClient) UpdateDashboard(ctx context.Context, id string, dashboard models.DashboardConfig, authToken string) error {
	// PATCH /api/metadata/dashboards/:id
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/metadata/dashboards/%s", id), dashboard, nil, authToken)
}

// DeleteDashboard deletes a dashboard configuration
func (c *NexusClient) DeleteDashboard(ctx context.Context, id string, authToken string) error {
	// DELETE /api/metadata/dashboards/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/metadata/dashboards/%s", id), nil, nil, authToken)
}

// GetRecycleBinItems retrieves items from the recycle bin
func (c *NexusClient) GetRecycleBinItems(ctx context.Context, scope string, authToken string) ([]models.RecycleBinItem, error) {
	// GET /api/data/recyclebin/items?scope=...
	var respMap map[string][]models.RecycleBinItem
	path := fmt.Sprintf("/api/data/recyclebin/items?scope=%s", scope)
	if err := c.doRequest(ctx, "GET", path, nil, &respMap, authToken); err != nil {
		return nil, err
	}
	if items, ok := respMap["data"]; ok {
		return items, nil
	}
	return nil, fmt.Errorf("invalid response format for recycle bin")
}

// RestoreRecord restores a record from the recycle bin
func (c *NexusClient) RestoreRecord(ctx context.Context, id string, authToken string) error {
	// POST /api/data/recyclebin/restore/:id
	return c.doRequest(ctx, "POST", fmt.Sprintf("/api/data/recyclebin/restore/%s", id), nil, nil, authToken)
}

// PurgeRecord permanently deletes a record from the recycle bin
func (c *NexusClient) PurgeRecord(ctx context.Context, id string, authToken string) error {
	// DELETE /api/data/recyclebin/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/data/recyclebin/%s", id), nil, nil, authToken)
}

// ExecuteFlow triggers a flow to run immediately
func (c *NexusClient) ExecuteFlow(ctx context.Context, flowID string, authToken string) error {
	// POST /api/flows/:flowId/execute
	return c.doRequest(ctx, "POST", fmt.Sprintf("/api/flows/%s/execute", flowID), nil, nil, authToken)
}

// CreateValidationRule creates a new validation rule
func (c *NexusClient) CreateValidationRule(ctx context.Context, rule models.ValidationRule, authToken string) (string, error) {
	// POST /api/metadata/validation-rules
	var rawResp map[string]interface{}
	if err := c.doRequest(ctx, "POST", "/api/metadata/validation-rules", rule, &rawResp, authToken); err != nil {
		return "", err
	}

	// Check "data" -> "id"
	if dataVal, ok := rawResp["data"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if id, ok := dataMap["id"].(string); ok {
				return id, nil
			}
		}
	}

	return "", fmt.Errorf("created rule missing ID")
}

// UpdateValidationRule updates an existing validation rule
func (c *NexusClient) UpdateValidationRule(ctx context.Context, id string, rule models.ValidationRule, authToken string) error {
	// PATCH /api/metadata/validation-rules/:id
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/metadata/validation-rules/%s", id), rule, nil, authToken)
}

// DeleteValidationRule deletes a validation rule
func (c *NexusClient) DeleteValidationRule(ctx context.Context, id string, authToken string) error {
	// DELETE /api/metadata/validation-rules/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/metadata/validation-rules/%s", id), nil, nil, authToken)
}
