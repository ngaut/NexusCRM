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
		return fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBytes))
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

	if objs, ok := respMap["schemas"]; ok {
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
	if obj, ok := respMap["schema"]; ok {
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

	if records, ok := respMap["records"]; ok {
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

	// Check "id" at top level
	if id, ok := rawResp["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("created record missing ID")
}

func (c *NexusClient) UpdateRecord(ctx context.Context, objectName, id string, data map[string]interface{}, authToken string) error {
	// PATCH /api/data/:objectApiName/:id
	return c.doRequest(ctx, "PATCH", fmt.Sprintf("/api/data/%s/%s", objectName, id), data, nil, authToken)
}

func (c *NexusClient) DeleteRecord(ctx context.Context, objectName, id string, authToken string) error {
	// DELETE /api/data/:objectApiName/:id
	return c.doRequest(ctx, "DELETE", fmt.Sprintf("/api/data/%s/%s", objectName, id), nil, nil, authToken)
}

// CreateDashboard creates a dashboard using the dedicated dashboard API
func (c *NexusClient) CreateDashboard(ctx context.Context, dashboard models.DashboardCreate, authToken string) (string, error) {
	// POST /api/metadata/dashboards
	var rawResp map[string]interface{}
	if err := c.doRequest(ctx, "POST", "/api/metadata/dashboards", dashboard, &rawResp, authToken); err != nil {
		return "", err
	}

	// Check "dashboard" -> "id"
	if dataVal, ok := rawResp["dashboard"]; ok {
		if dataMap, ok := dataVal.(map[string]interface{}); ok {
			if id, ok := dataMap["id"].(string); ok {
				return id, nil
			}
		}
	}

	// Check "id" at top level
	if id, ok := rawResp["id"].(string); ok {
		return id, nil
	}

	return "", fmt.Errorf("created dashboard missing ID")
}

// CreateObject creates a new object schema
func (c *NexusClient) CreateObject(ctx context.Context, schema models.ObjectMetadata, authToken string) error {
	// POST /api/metadata/objects
	return c.doRequest(ctx, "POST", "/api/metadata/objects", schema, nil, authToken)
}

// CreateField creates a new field on an object
func (c *NexusClient) CreateField(ctx context.Context, objectName string, field models.FieldMetadata, authToken string) error {
	// POST /api/metadata/objects/:apiName/fields
	return c.doRequest(ctx, "POST", fmt.Sprintf("/api/metadata/objects/%s/fields", objectName), field, nil, authToken)
}
