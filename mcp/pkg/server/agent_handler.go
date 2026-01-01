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

// GetConversation returns a conversation for the current user
// If id query param provided, loads that specific conversation
// Otherwise loads the active (most recent) conversation
func (h *AgentHandler) GetConversation(c *gin.Context) {
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)

	convID := c.Query("id")
	var record models.SObject

	if convID != "" {
		// Load specific conversation by ID
		record, err = nexusClient.GetRecord(c.Request.Context(), "_System_AI_Conversation", convID, token)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
			return
		}
		// Verify ownership
		if record["user_id"] != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your conversation"})
			return
		}
	} else {
		// Load most recent active conversation
		queryReq := models.QueryRequest{
			ObjectAPIName: "_System_AI_Conversation",
			FilterExpr:    fmt.Sprintf("user_id == '%s' && is_active == true", user.ID),
			Limit:         1,
		}
		records, err := nexusClient.Query(c.Request.Context(), queryReq, token)
		if err != nil || len(records) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"conversation": nil,
				"messages":     []interface{}{},
			})
			return
		}
		record = records[0]
	}

	// Parse messages from JSON
	var messages []interface{}
	if msgData, ok := record["messages"]; ok && msgData != nil {
		switch v := msgData.(type) {
		case []interface{}:
			messages = v
		case string:
			if err := json.Unmarshal([]byte(v), &messages); err != nil {
				messages = []interface{}{}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": gin.H{
			"id":    record["id"],
			"title": record["title"],
		},
		"messages": messages,
	})
}

// SaveConversation saves/updates a conversation for the current user
// If conversation_id provided, updates that conversation
// If no conversation_id, creates new conversation (sets as active)
func (h *AgentHandler) SaveConversation(c *gin.Context) {
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		ConversationID string        `json:"conversation_id"`
		Messages       []interface{} `json:"messages"`
		Title          string        `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)

	messagesJSON, _ := json.Marshal(req.Messages)

	if req.ConversationID != "" {
		// Update existing conversation
		// Verify ownership first
		record, err := nexusClient.GetRecord(c.Request.Context(), "_System_AI_Conversation", req.ConversationID, token)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
			return
		}
		if record["user_id"] != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "not your conversation"})
			return
		}

		updateData := map[string]interface{}{
			"messages": string(messagesJSON),
		}
		if req.Title != "" {
			updateData["title"] = req.Title
		}
		err = nexusClient.UpdateRecord(c.Request.Context(), "_System_AI_Conversation", req.ConversationID, updateData, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": req.ConversationID, "status": "updated"})
	} else {
		// Create new conversation
		// First, deactivate any currently active conversations
		queryReq := models.QueryRequest{
			ObjectAPIName: "_System_AI_Conversation",
			FilterExpr:    fmt.Sprintf("user_id == '%s' && is_active == true", user.ID),
			Limit:         10,
		}
		activeRecords, _ := nexusClient.Query(c.Request.Context(), queryReq, token)
		for _, rec := range activeRecords {
			if id, ok := rec["id"].(string); ok {
				nexusClient.UpdateRecord(c.Request.Context(), "_System_AI_Conversation", id, map[string]interface{}{
					"is_active": false,
				}, token)
			}
		}

		// Create new active conversation
		title := req.Title
		if title == "" && len(req.Messages) > 0 {
			// Auto-generate title from first user message
			for _, msg := range req.Messages {
				if msgMap, ok := msg.(map[string]interface{}); ok {
					if msgMap["role"] == "user" {
						if content, ok := msgMap["content"].(string); ok {
							title = content
							if len(title) > 50 {
								title = title[:47] + "..."
							}
							break
						}
					}
				}
			}
		}
		if title == "" {
			title = "New Conversation"
		}

		createData := map[string]interface{}{
			"user_id":   user.ID,
			"title":     title,
			"messages":  string(messagesJSON),
			"is_active": true,
		}
		id, err := nexusClient.CreateRecord(c.Request.Context(), "_System_AI_Conversation", createData, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id, "status": "created"})
	}
}

// ClearConversation clears the current user's conversation
func (h *AgentHandler) ClearConversation(c *gin.Context) {
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)

	// Find and delete the active conversation
	queryReq := models.QueryRequest{
		ObjectAPIName: "_System_AI_Conversation",
		FilterExpr:    fmt.Sprintf("user_id == '%s' && is_active == true", user.ID),
		Limit:         1,
	}
	records, _ := nexusClient.Query(c.Request.Context(), queryReq, token)

	if len(records) > 0 {
		convID, ok := records[0]["id"].(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid conversation id"})
			return
		}
		// Clear messages instead of delete (keeps history record)
		updateData := map[string]interface{}{
			"messages": "[]",
		}
		nexusClient.UpdateRecord(c.Request.Context(), "_System_AI_Conversation", convID, updateData, token)
	}

	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}

// ListConversations returns all conversations for the current user
func (h *AgentHandler) ListConversations(c *gin.Context) {
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)

	queryReq := models.QueryRequest{
		ObjectAPIName: "_System_AI_Conversation",
		FilterExpr:    fmt.Sprintf("user_id == '%s'", user.ID),
		SortField:     "last_modified_date",
		SortDirection: "desc",
		Limit:         100,
	}

	records, err := nexusClient.Query(c.Request.Context(), queryReq, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"conversations": []interface{}{}})
		return
	}

	// Map to simplified response
	conversations := make([]gin.H, 0, len(records))
	for _, record := range records {
		conv := gin.H{
			"id":                 record["id"],
			"title":              record["title"],
			"is_active":          record["is_active"],
			"created_date":       record["created_date"],
			"last_modified_date": record["last_modified_date"],
		}
		conversations = append(conversations, conv)
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// DeleteConversation deletes a specific conversation
func (h *AgentHandler) DeleteConversation(c *gin.Context) {
	user := h.userExtractor(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	token, err := h.getAuthToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	convID := c.Param("id")
	if convID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id required"})
		return
	}

	apiBaseURL := "http://localhost:3001"
	if url := os.Getenv("API_BASE_URL"); url != "" {
		apiBaseURL = url
	}
	nexusClient := client.NewNexusClient(apiBaseURL)

	// Verify ownership before delete
	record, err := nexusClient.GetRecord(c.Request.Context(), "_System_AI_Conversation", convID, token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if record["user_id"] != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your conversation"})
		return
	}

	// Delete the conversation
	err = nexusClient.DeleteRecord(c.Request.Context(), "_System_AI_Conversation", convID, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
