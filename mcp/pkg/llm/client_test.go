package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIClient_Chat(t *testing.T) {
	// 1. Create Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path /v1/chat/completions, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		// Decode Body
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if len(req.Messages) != 1 || req.Messages[0].Content != "Hello" {
			t.Errorf("Unexpected messages: %v", req.Messages)
		}

		// Return Mock Response
		resp := Response{
			ID: "mock-123",
			Choices: []Choice{{
				Message: Message{
					Role:    "assistant",
					Content: "Hi there!",
				},
			}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 2. Initialize Client
	client := NewOpenAIClient(server.URL+"/v1/chat/completions", "test-key")

	// 3. Execute
	resp, err := client.Chat(context.Background(), Request{
		Messages: []Message{{Role: "user", Content: "Hello"}},
	})

	// 4. Verify
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hi there!" {
		t.Errorf("Expected 'Hi there!', got '%s'", resp.Choices[0].Message.Content)
	}
}
