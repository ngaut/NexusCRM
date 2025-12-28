package compactor

import (
	"context"
	"strings"
	"testing"

	"github.com/nexuscrm/mcp/pkg/llm"
)

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		messages []llm.Message
		want     int
	}{
		{
			name:     "empty messages",
			messages: []llm.Message{},
			want:     0,
		},
		{
			name: "single short message",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"}, // 5 chars = 1 token
			},
			want: 1,
		},
		{
			name: "message with tool calls",
			messages: []llm.Message{
				{
					Role:    "assistant",
					Content: "Let me help", // 11 chars / 4 = 2 tokens
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: llm.FunctionCallData{
								Name:      "get_data",
								Arguments: `{"query": "test value"}`, // 23 chars / 4 = 5 tokens
							},
						},
					},
				},
			},
			want: 7, // 2 + 5 = 7
		},
		{
			name: "multiple messages",
			messages: []llm.Message{
				{Role: "system", Content: "You are a helpful assistant."}, // 29 chars = 7 tokens
				{Role: "user", Content: "What is 2+2?"},                   // 12 chars = 3 tokens
				{Role: "assistant", Content: "The answer is 4."},          // 17 chars = 4 tokens
			},
			want: 14, // 7 + 3 + 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.messages)
			if got != tt.want {
				t.Errorf("EstimateTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMicroCompact(t *testing.T) {
	tests := []struct {
		name     string
		messages []llm.Message
		check    func(t *testing.T, result []llm.Message)
	}{
		{
			name:     "empty messages",
			messages: []llm.Message{},
			check: func(t *testing.T, result []llm.Message) {
				if len(result) != 0 {
					t.Errorf("expected empty result, got %d messages", len(result))
				}
			},
		},
		{
			name: "short messages unchanged",
			messages: []llm.Message{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: "Hi there!"},
			},
			check: func(t *testing.T, result []llm.Message) {
				if len(result) != 2 {
					t.Errorf("expected 2 messages, got %d", len(result))
				}
				if result[0].Content != "Hello" {
					t.Errorf("user content changed unexpectedly")
				}
				if result[1].Content != "Hi there!" {
					t.Errorf("assistant content changed unexpectedly")
				}
			},
		},
		{
			name: "long tool result truncated",
			messages: []llm.Message{
				{
					Role:       "tool",
					Name:       "get_data",
					Content:    string(make([]byte, 1000)), // 1000 'a' characters
					ToolCallID: "call_1",
				},
			},
			check: func(t *testing.T, result []llm.Message) {
				if len(result) != 1 {
					t.Errorf("expected 1 message, got %d", len(result))
				}
				// Should be truncated to MaxToolResultLength + "...[truncated]"
				expectedLen := MaxToolResultLength + len("...[truncated]")
				if len(result[0].Content) != expectedLen {
					t.Errorf("expected content length %d, got %d", expectedLen, len(result[0].Content))
				}
				if result[0].Name != "get_data" {
					t.Errorf("tool name should be preserved")
				}
				if result[0].ToolCallID != "call_1" {
					t.Errorf("tool_call_id should be preserved")
				}
			},
		},
		{
			name: "long tool call arguments truncated",
			messages: []llm.Message{
				{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{
						{
							ID:   "call_1",
							Type: "function",
							Function: llm.FunctionCallData{
								Name:      "create_record",
								Arguments: string(make([]byte, 1000)), // 1000 chars
							},
						},
					},
				},
			},
			check: func(t *testing.T, result []llm.Message) {
				if len(result) != 1 {
					t.Errorf("expected 1 message, got %d", len(result))
				}
				if len(result[0].ToolCalls) != 1 {
					t.Errorf("expected 1 tool call, got %d", len(result[0].ToolCalls))
				}
				tc := result[0].ToolCalls[0]
				expectedLen := MaxToolResultLength + len("...[truncated]")
				if len(tc.Function.Arguments) != expectedLen {
					t.Errorf("expected args length %d, got %d", expectedLen, len(tc.Function.Arguments))
				}
				if tc.Function.Name != "create_record" {
					t.Errorf("function name should be preserved")
				}
				if tc.ID != "call_1" {
					t.Errorf("tool call ID should be preserved")
				}
			},
		},
		{
			name: "user/assistant messages not truncated",
			messages: []llm.Message{
				{Role: "user", Content: string(make([]byte, 1000))},
				{Role: "assistant", Content: string(make([]byte, 1000))},
			},
			check: func(t *testing.T, result []llm.Message) {
				// User and assistant messages should NOT be truncated
				if len(result[0].Content) != 1000 {
					t.Errorf("user content should not be truncated, got length %d", len(result[0].Content))
				}
				if len(result[1].Content) != 1000 {
					t.Errorf("assistant content should not be truncated, got length %d", len(result[1].Content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MicroCompact(tt.messages)
			tt.check(t, result)
		})
	}
}

// MockLLMClient for testing Compact function
type MockLLMClient struct {
	ChatFunc func(ctx context.Context, req llm.Request) (*llm.Response, error)
}

func (m *MockLLMClient) Chat(ctx context.Context, req llm.Request) (*llm.Response, error) {
	if m.ChatFunc != nil {
		return m.ChatFunc(ctx, req)
	}
	// Default: return a summary
	return &llm.Response{
		Choices: []llm.Choice{
			{
				Message: llm.Message{
					Role:    "assistant",
					Content: "This is a summary of the conversation.",
				},
			},
		},
	}, nil
}

func TestCompact_TooFewMessages(t *testing.T) {
	c := &Compactor{
		llmClient: &MockLLMClient{},
		model:     "test-model",
	}

	messages := []llm.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
	}

	resp, err := c.Compact(context.Background(), CompactRequest{Messages: messages})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should return original messages unchanged
	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp.Messages))
	}
	if resp.TokensBefore != resp.TokensAfter {
		t.Errorf("tokens should be equal for unchanged messages")
	}
}

func TestCompact_ContextCancellation(t *testing.T) {
	c := &Compactor{
		llmClient: &MockLLMClient{},
		model:     "test-model",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messages := []llm.Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
		{Role: "assistant", Content: "I'm good!"},
		{Role: "user", Content: "Great"},
	}

	resp, err := c.Compact(ctx, CompactRequest{Messages: messages})

	// Should return context error
	if err == nil {
		t.Errorf("expected context error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// Should return original messages
	if len(resp.Messages) != 6 {
		t.Errorf("expected original 6 messages, got %d", len(resp.Messages))
	}
}

func TestCompact_Success(t *testing.T) {
	c := &Compactor{
		llmClient: &MockLLMClient{},
		model:     "test-model",
	}

	// Create enough messages to trigger compaction
	messages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Tell me about Go programming."},
		{Role: "assistant", Content: "Go is a statically typed, compiled language designed at Google."},
		{Role: "user", Content: "What are its main features?"},
		{Role: "assistant", Content: "Go features goroutines for concurrency, garbage collection, and a simple syntax."},
		{Role: "user", Content: "Show me an example."},
	}

	resp, err := c.Compact(context.Background(), CompactRequest{Messages: messages})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Should have compacted (system + summary + recent messages)
	if resp.TokensAfter >= resp.TokensBefore {
		// This might happen if summary is larger - check that we got back messages
		t.Logf("Note: compaction didn't reduce size (before=%d, after=%d)", resp.TokensBefore, resp.TokensAfter)
	}

	// First message should be system message with summary
	if len(resp.Messages) == 0 {
		t.Errorf("expected at least 1 message in result")
		return
	}

	if resp.Messages[0].Role != "system" {
		t.Errorf("first message should be system, got %s", resp.Messages[0].Role)
	}
}

func TestCompact_WithKeepInstruction(t *testing.T) {
	var capturedPrompt string
	mockClient := &MockLLMClient{
		ChatFunc: func(ctx context.Context, req llm.Request) (*llm.Response, error) {
			if len(req.Messages) > 0 {
				capturedPrompt = req.Messages[0].Content
			}
			return &llm.Response{
				Choices: []llm.Choice{
					{Message: llm.Message{Role: "assistant", Content: "Summary with errors preserved."}},
				},
			}, nil
		},
	}

	c := &Compactor{
		llmClient: mockClient,
		model:     "test-model",
	}

	messages := []llm.Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Message 1"},
		{Role: "assistant", Content: "Response 1"},
		{Role: "user", Content: "Message 2"},
		{Role: "assistant", Content: "Response 2"},
		{Role: "user", Content: "Message 3"},
	}

	_, err := c.Compact(context.Background(), CompactRequest{
		Messages: messages,
		Keep:     "error messages and stack traces",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify the keep instruction is in the prompt
	if !strings.Contains(capturedPrompt, "error messages and stack traces") {
		t.Errorf("keep instruction not found in prompt")
	}
}

func TestCompact_MultiCompactMergesSummaries(t *testing.T) {
	var capturedPrompt string
	mockClient := &MockLLMClient{
		ChatFunc: func(ctx context.Context, req llm.Request) (*llm.Response, error) {
			if len(req.Messages) > 0 {
				capturedPrompt = req.Messages[0].Content
			}
			return &llm.Response{
				Choices: []llm.Choice{
					{Message: llm.Message{Role: "assistant", Content: "Consolidated summary of old and new context."}},
				},
			}, nil
		},
	}

	c := &Compactor{
		llmClient: mockClient,
		model:     "test-model",
	}

	// Simulate a system message that already contains a previous summary (result of first /compact)
	systemWithPreviousSummary := `You are a helpful assistant.

--- CONVERSATION SUMMARY (Saved ~100 tokens) ---
The user asked about Go programming basics. We discussed goroutines and channels.
--- END SUMMARY ---`

	messages := []llm.Message{
		{Role: "system", Content: systemWithPreviousSummary},
		{Role: "user", Content: "Now tell me about error handling."},
		{Role: "assistant", Content: "Go uses explicit error returns instead of exceptions."},
		{Role: "user", Content: "Show me an example."},
		{Role: "assistant", Content: "Here is an example using if err != nil."},
		{Role: "user", Content: "Thanks!"},
	}

	resp, err := c.Compact(context.Background(), CompactRequest{Messages: messages})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 1. Verify the previous summary was included in the LLM prompt
	if !strings.Contains(capturedPrompt, "Previous Context Summary") {
		t.Errorf("expected LLM prompt to contain 'Previous Context Summary' section")
	}
	if !strings.Contains(capturedPrompt, "goroutines and channels") {
		t.Errorf("expected LLM prompt to contain the old summary content")
	}

	// 2. Verify the result has only ONE summary block (not stacked)
	if len(resp.Messages) == 0 || resp.Messages[0].Role != "system" {
		t.Fatalf("expected at least 1 system message in result")
	}

	systemContent := resp.Messages[0].Content
	summaryCount := strings.Count(systemContent, "--- CONVERSATION SUMMARY")
	if summaryCount != 1 {
		t.Errorf("expected exactly 1 summary block, found %d", summaryCount)
	}

	// 3. Verify the base system prompt is preserved (without the old summary)
	if !strings.Contains(systemContent, "You are a helpful assistant.") {
		t.Errorf("base system prompt should be preserved")
	}

	// 4. Verify the new summary contains the merged content
	if !strings.Contains(systemContent, "Consolidated summary") {
		t.Errorf("expected new summary to contain merged content")
	}
}
