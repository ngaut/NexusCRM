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

	// Test case 2: "record" envelope (Backend Standard)
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Success",
			"record": map[string]interface{}{
				"id": "record-id-456",
			},
		})
	}))
	defer server2.Close()

	client2 := NewNexusClient(server2.URL)
	id2, err := client2.CreateRecord(context.Background(), "Account", map[string]interface{}{"name": "Another Account"}, "test-token")
	assert.NoError(t, err)
	assert.Equal(t, "record-id-456", id2)
}

func TestSearchObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/data/search/Account", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Acme", r.URL.Query().Get("term"))

		response := map[string]interface{}{
			"records": []models.SObject{
				{"id": "1", "name": "Acme Corp"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNexusClient(server.URL)
	results, err := client.SearchObject(context.Background(), "Account", "Acme", "test-token")

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Acme Corp", results[0]["name"])
}

func TestGetRecord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/data/Account/123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		response := map[string]interface{}{
			"record": models.SObject{
				"id": "123", "name": "Acme Corp",
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewNexusClient(server.URL)
	record, err := client.GetRecord(context.Background(), "Account", "123", "test-token")

	assert.NoError(t, err)
	assert.Equal(t, "123", record["id"])
	assert.Equal(t, "Acme Corp", record["name"])
}
