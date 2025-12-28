package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/nexuscrm/mcp/pkg/agent"
	"github.com/nexuscrm/mcp/pkg/client"
	"github.com/nexuscrm/mcp/pkg/compactor"
	"github.com/nexuscrm/mcp/pkg/contextstore"
	"github.com/nexuscrm/mcp/pkg/llm"
	"github.com/nexuscrm/mcp/pkg/mcp"
	"github.com/nexuscrm/mcp/pkg/models"
)

type AgentHandler struct {
	agentSvc      *agent.AgentService
	compactor     *compactor.Compactor
	contextStore  *contextstore.ContextStore
	userExtractor func(c *gin.Context) *models.UserSession
}

func NewAgentHandler(userExtractor func(c *gin.Context) *models.UserSession, contextStore *contextstore.ContextStore) *AgentHandler {
	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1234/v1/chat/completions" // Default to LM Studio usually
	}
	llmClient := llm.NewOpenAIClient(baseURL, os.Getenv("LLM_API_KEY"))

	// Create Agent Service
	// ToolBus is now decoupled and uses NexusClient
	apiBaseURL := "http://localhost:3001" // Default to local
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)
	toolBus := NewToolBusService(nexusClient, contextStore)

	agentSvc := agent.NewAgentService(llmClient, toolBus, contextStore)

	return &AgentHandler{
		agentSvc:      agentSvc,
		compactor:     agentSvc.GetCompactor(), // Reuse compactor from AgentService
		contextStore:  contextStore,
		userExtractor: userExtractor,
	}
}

// ChatStream handles SSE streaming for agent chat
func (h *AgentHandler) ChatStream(c *gin.Context) {
	var req agent.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract User using provided extractor
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	req.User = user

	// Extract Token (Cookie or Header)
	token, err := c.Cookie("auth_token")
	if err != nil || token == "" {
		// Try Header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	// Create Context with User AND Token for ToolBus (with cancellation)
	// Note: We use "user" and "auth_token" keys.
	// We pass the MCP user object itself into context if needed by tools,
	// though tools mostly rely on auth_token.
	ctxValue := context.WithValue(c.Request.Context(), mcp.ContextKeyUser, user)
	ctxValue = context.WithValue(ctxValue, mcp.ContextKeyAuthToken, token)
	ctx, cancel := context.WithCancel(ctxValue)
	defer cancel()

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Create event channel
	eventChan := make(chan agent.StreamEvent, 10)

	// Start streaming in goroutine
	go h.agentSvc.ChatStream(ctx, req, eventChan)

	// Stream events to client
	c.Stream(func(w io.Writer) bool {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return false
			}
			data, _ := json.Marshal(event)
			c.SSEvent("message", string(data))
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// GetContext returns the current context for the authenticated user
func (h *AgentHandler) GetContext(c *gin.Context) {
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	session := h.contextStore.GetSession(token)
	items := session.ListItems()
	totalTokens := session.GetTotalTokens()

	type ContextResponseItem struct {
		Path      string `json:"path"`
		TokenSize int    `json:"token_size"`
		Content   string `json:"content,omitempty"`
	}

	includeContent := c.Query("include_content") == "true"

	responseItems := make([]ContextResponseItem, len(items))
	for i, item := range items {
		respItem := ContextResponseItem{
			Path:      item.Path,
			TokenSize: item.TokenSize,
		}
		if includeContent {
			respItem.Content = item.Content
		}
		responseItems[i] = respItem
	}

	c.JSON(http.StatusOK, gin.H{
		"items":         responseItems,
		"total_tokens":  totalTokens,
		"system_prompt": h.agentSvc.GetBaseSystemPrompt(),
	})
}

// getAuthToken extracts the auth token from cookie or header
func (h *AgentHandler) getAuthToken(c *gin.Context) (string, error) {
	// Extract Token (Cookie or Header)
	token, err := c.Cookie("auth_token")
	if err != nil || token == "" {
		// Try Header
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}
	if token == "" {
		return "", fmt.Errorf("unauthorized")
	}
	return token, nil
}

// CompactContext compacts conversation history to reduce token usage
func (h *AgentHandler) CompactContext(c *gin.Context) {
	var req compactor.CompactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages required"})
		return
	}

	ctx := c.Request.Context()

	resp, err := h.compactor.Compact(ctx, req)
	if err != nil {
		// Still return the response (with original messages) but include error
		c.JSON(http.StatusOK, gin.H{
			"messages":      resp.Messages,
			"tokens_before": resp.TokensBefore,
			"tokens_after":  resp.TokensAfter,
			"warning":       err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages":      resp.Messages,
		"tokens_before": resp.TokensBefore,
		"tokens_after":  resp.TokensAfter,
	})
}
