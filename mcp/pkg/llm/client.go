package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
	// ToolCallID is required for tool response messages
	ToolCallID string `json:"tool_call_id,omitempty"`
	// FunctionCall support omitted in favor of standard 'tools' if supported,
	// but legacy models might use it. Sticking to standard Chat API structure.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	// Timestamp is when this message was created
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function FunctionCallData `json:"function"`
}

type FunctionCallData struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Request represents the LLM request payload
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

// Tool represents a tool definition passed to the LLM
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"` // JSON Schema
}

// Response represents the LLM response
type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client interface
type Client interface {
	Chat(ctx context.Context, req Request) (*Response, error)
}

// OpenAIClient implementation
type OpenAIClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewOpenAIClient(baseURL, apiKey string) *OpenAIClient {
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1/chat/completions" // Default to LM Studio usually
	}
	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (c *OpenAIClient) Chat(ctx context.Context, req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp Response
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp, nil
}
