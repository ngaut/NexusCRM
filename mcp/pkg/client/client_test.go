package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nexuscrm/mcp/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestListObjects(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/metadata/objects", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		response := map[string]interface{}{
			"schemas": []models.ObjectMetadata{
				{APIName: "Account", Label: "Account"},
				{APIName: "Contact", Label: "Contact"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNexusClient(server.URL)
	objects, err := client.ListObjects(context.Background(), "test-token")

	assert.NoError(t, err)
	assert.Len(t, objects, 2)
	assert.Equal(t, "Account", objects[0].APIName)
}

func TestQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/data/query", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req models.QueryRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Account", req.ObjectAPIName)
		assert.NotEmpty(t, req.FilterExpr)

		response := map[string]interface{}{
			"records": []models.SObject{
				{"id": "1", "name": "Acme Corp"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNexusClient(server.URL)
	req := models.QueryRequest{
		ObjectAPIName: "Account",
		FilterExpr:    "name == 'Acme'",
	}
	results, err := client.Query(context.Background(), req, "test-token")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Acme Corp", results[0]["name"])
}

func TestCreateRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/data/Account", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "New Account", body["name"])

		// Return ID envelope
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "new-id-123",
		})
	}))
	defer server.Close()

	client := NewNexusClient(server.URL)
	id, err := client.CreateRecord(context.Background(), "Account", map[string]interface{}{"name": "New Account"}, "test-token")

	assert.NoError(t, err)
	assert.Equal(t, "new-id-123", id)
}
