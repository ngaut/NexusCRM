package compactor

import "github.com/nexuscrm/mcp/pkg/llm"

// CompactRequest represents a request to compact conversation history
type CompactRequest struct {
	Messages []llm.Message `json:"messages"`
	Keep     string        `json:"keep,omitempty"` // Optional instruction for what to preserve
}

// CompactResponse represents the result of context compaction
type CompactResponse struct {
	Messages     []llm.Message `json:"messages"`
	TokensBefore int           `json:"tokens_before"`
	TokensAfter  int           `json:"tokens_after"`
}
