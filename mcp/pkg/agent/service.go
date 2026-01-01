package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nexuscrm/mcp/pkg/compactor"
	"github.com/nexuscrm/mcp/pkg/contextstore"
	"github.com/nexuscrm/mcp/pkg/llm"
	"github.com/nexuscrm/mcp/pkg/mcp"
	"github.com/nexuscrm/mcp/pkg/models"
)

type ToolBus interface {
	HandleListTools(ctx context.Context, params json.RawMessage) (interface{}, error)
	HandleCallTool(ctx context.Context, params json.RawMessage) (interface{}, error)
}

// Auto-compact configuration defaults
const (
	DefaultMaxContextTokens     = 100000 // Default max tokens
	DefaultAutoCompactThreshold = 0.75   // Compact at 75% capacity
)

type AgentService struct {
	llmClient         llm.Client
	toolBus           ToolBus
	contextStore      *contextstore.ContextStore
	compactor         *compactor.Compactor
	maxContextTokens  int
	autoCompactThresh float64
}

func NewAgentService(llmClient llm.Client, toolBus ToolBus, contextStore *contextstore.ContextStore) *AgentService {
	// Read config from environment
	maxTokens := DefaultMaxContextTokens
	if val := os.Getenv("MAX_CONTEXT_TOKENS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxTokens = parsed
		}
	}

	threshold := DefaultAutoCompactThreshold
	if val := os.Getenv("AUTO_COMPACT_THRESHOLD"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			threshold = parsed
		}
	}

	return &AgentService{
		llmClient:         llmClient,
		toolBus:           toolBus,
		contextStore:      contextStore,
		compactor:         compactor.NewCompactor(llmClient),
		maxContextTokens:  maxTokens,
		autoCompactThresh: threshold,
	}
}

func (s *AgentService) GetCompactor() *compactor.Compactor {
	return s.compactor
}

// GetBaseSystemPrompt returns the current system prompt with dynamic date
func (s *AgentService) GetBaseSystemPrompt() string {
	systemPrompt := fmt.Sprintf("You are Nexus, an AI assistant for NexusCRM. Today is %s.\n", time.Now().Format("Monday, January 2, 2006"))
	systemPrompt += "\nPRINCIPLES:"
	systemPrompt += "\n1. EXPLORE BEFORE ACTING - Like exploring a new codebase, first understand what's available, then dive into specifics."
	systemPrompt += "\n2. TREE EXPLORATION - Start broad (list all), then narrow down (get details), then act (CRUD). Don't try to do everything at once."
	systemPrompt += "\n\nYou have access to a dynamic CRM system. Objects and fields are metadata-driven."
	systemPrompt += " Think step by step. If a tool fails, read the error and adapt."
	return systemPrompt
}

// ChatRequest represents a single turn of chat
type ChatRequest struct {
	Model    string        `json:"model"` // Optional, override default
	Messages []llm.Message `json:"messages"`
	User     *models.UserSession
}

// ChatResponse final response to the UI
type ChatResponse struct {
	Content string        `json:"content"`
	History []llm.Message `json:"history"` // Full history including tool calls for debugging/UI
}

// StreamEventType represents different stages of agent processing
type StreamEventType string

const (
	EventThinking    StreamEventType = "thinking"
	EventToolCall    StreamEventType = "tool_call"
	EventToolResult  StreamEventType = "tool_result"
	EventContent     StreamEventType = "content"
	EventDone        StreamEventType = "done"
	EventError       StreamEventType = "error"
	EventAutoCompact StreamEventType = "auto_compact" // Context was automatically compacted
)

// StreamEvent represents a single streaming event
type StreamEvent struct {
	Type         StreamEventType `json:"type"`
	Content      string          `json:"content,omitempty"`
	ToolName     string          `json:"tool_name,omitempty"`
	ToolCallID   string          `json:"tool_call_id,omitempty"`
	ToolArgs     string          `json:"tool_args,omitempty"`
	ToolResult   string          `json:"tool_result,omitempty"`
	IsError      bool            `json:"is_error,omitempty"`
	History      []llm.Message   `json:"history,omitempty"`
	TokensBefore int             `json:"tokens_before,omitempty"` // For auto_compact events
	TokensAfter  int             `json:"tokens_after,omitempty"`  // For auto_compact events
}

// ChatStream processes a chat request and streams events to the provided channel
func (s *AgentService) ChatStream(ctx context.Context, req ChatRequest, eventChan chan<- StreamEvent) {
	defer close(eventChan)

	// Helper to send event
	emit := func(event StreamEvent) {
		select {
		case eventChan <- event:
		case <-ctx.Done():
			return
		}
	}

	// 1. Prepare Tools
	// emit(StreamEvent{Type: EventThinking, Content: "Preparing tools..."})

	listRespUntyped, err := s.toolBus.HandleListTools(ctx, []byte("{}"))
	if err != nil {
		emit(StreamEvent{Type: EventError, Content: fmt.Sprintf("Failed to list tools: %v", err), IsError: true})
		return
	}

	listResp, ok := listRespUntyped.(mcp.ListToolsResult)
	if !ok {
		emit(StreamEvent{Type: EventError, Content: "Unexpected tool list format", IsError: true})
		return
	}

	var tools []llm.Tool
	for _, t := range listResp.Tools {
		tools = append(tools, llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	// 2. Initialize Conversation
	model := req.Model
	if model == "" {
		model = "nvidia-nemotron-3-nano-30b-a3b-mlx"
	}

	// 3. Inject Context from ContextStore
	authToken, _ := ctx.Value(mcp.ContextKeyAuthToken).(string)
	contextInjection := ""
	if authToken != "" {
		session := s.contextStore.GetSession(authToken)
		items := session.ListItems()
		if len(items) > 0 {
			contextInjection = "\n\nACTIVE CONTEXT FILES (Priority over general knowledge):\n"
			for _, item := range items {
				contextInjection += fmt.Sprintf("\n--- FILE: %s ---\n%s\n--- END FILE ---\n", item.Path, item.Content)
			}
		}
	}

	// 4. Prepare Messages
	messages := req.Messages
	if len(messages) == 0 || messages[0].Role != "system" {
		// No system prompt provided, generate default
		systemPrompt := s.GetBaseSystemPrompt()
		systemPrompt += contextInjection // Append context

		now := time.Now()
		messages = append([]llm.Message{{Role: "system", Content: systemPrompt, Timestamp: &now}}, messages...)
	} else {
		// System prompt exists, append context if not already present
		if contextInjection != "" && !strings.Contains(messages[0].Content, "ACTIVE CONTEXT FILES") {
			messages[0].Content += contextInjection
		}
	}

	// 4. Auto-Compact Check - Compact if token count exceeds threshold
	tokenCount := compactor.EstimateTokens(messages)
	threshold := int(float64(s.maxContextTokens) * s.autoCompactThresh)
	if tokenCount > threshold {
		compactReq := compactor.CompactRequest{
			Messages: messages,
		}
		compactResp, err := s.compactor.Compact(ctx, compactReq)
		if err == nil && compactResp.TokensAfter < compactResp.TokensBefore {
			// Successful compaction
			messages = compactResp.Messages
			emit(StreamEvent{
				Type:         EventAutoCompact,
				Content:      fmt.Sprintf("Context auto-compacted: %d → %d tokens", compactResp.TokensBefore, compactResp.TokensAfter),
				TokensBefore: compactResp.TokensBefore,
				TokensAfter:  compactResp.TokensAfter,
			})
		}
		// If compaction fails or doesn't reduce size, continue with original messages
	}

	// Max agent reasoning steps (not tool count, but LLM call iterations)
	maxAgentSteps := 100

	// ReAct Loop with Streaming
	for i := 0; i < maxAgentSteps; i++ {
		llmReq := llm.Request{
			Model:       model,
			Messages:    messages,
			Tools:       tools,
			Temperature: 0.7,
		}

		resp, err := s.llmClient.Chat(ctx, llmReq)
		if err != nil {
			emit(StreamEvent{Type: EventError, Content: fmt.Sprintf("LLM Error: %v", err), IsError: true})
			return
		}

		if len(resp.Choices) == 0 {
			emit(StreamEvent{Type: EventError, Content: "Empty response from LLM", IsError: true})
			return
		}

		choice := resp.Choices[0]
		assistantMsg := choice.Message
		// Add timestamp to assistant message
		now := time.Now()
		assistantMsg.Timestamp = &now
		messages = append(messages, assistantMsg)

		if len(assistantMsg.ToolCalls) > 0 {
			for _, tc := range assistantMsg.ToolCalls {
				// Emit tool call event
				emit(StreamEvent{
					Type:       EventToolCall,
					ToolName:   tc.Function.Name,
					ToolCallID: tc.ID,
					ToolArgs:   tc.Function.Arguments,
				})

				// Execute tool
				callParams := map[string]interface{}{
					"name":      tc.Function.Name,
					"arguments": json.RawMessage(tc.Function.Arguments),
				}
				callParamsBytes, _ := json.Marshal(callParams)

				resultUntyped, err := s.toolBus.HandleCallTool(ctx, callParamsBytes)

				var toolOutput string
				var isError bool
				if err != nil {
					toolOutput = fmt.Sprintf("Error: %v", err)
					isError = true
				} else {
					result, ok := resultUntyped.(mcp.CallToolResult)
					if ok {
						for _, content := range result.Content {
							if content.Type == "text" {
								toolOutput += content.Text + "\n"
							}
						}
						isError = result.IsError
					} else {
						toolOutput = fmt.Sprintf("Unexpected format: %v", resultUntyped)
						isError = true
					}
				}

				// Emit tool result
				emit(StreamEvent{
					Type:       EventToolResult,
					ToolName:   tc.Function.Name,
					ToolCallID: tc.ID,
					ToolResult: toolOutput,
					IsError:    isError,
				})

				toolNow := time.Now()
				messages = append(messages, llm.Message{
					Role:       "tool",
					Content:    toolOutput,
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Timestamp:  &toolNow,
				})
			}
		} else {
			// Final answer
			emit(StreamEvent{
				Type:    EventContent,
				Content: assistantMsg.Content,
			})
			emit(StreamEvent{
				Type:    EventDone,
				History: messages,
			})
			return
		}
	}

	// Max iterations
	emit(StreamEvent{
		Type:    EventContent,
		Content: "I apologize, but I was unable to complete the request within the step limit. Please try a simpler request.",
	})
	emit(StreamEvent{
		Type:    EventDone,
		History: messages,
	})
}
