package compactor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nexuscrm/mcp/pkg/llm"
)

// DefaultSummarizationModel is the default model used for summarization
const DefaultSummarizationModel = "nvidia-nemotron-3-nano-30b-a3b-mlx"

// Compactor provides conversation history compaction services
type Compactor struct {
	llmClient llm.Client
	model     string
}

// NewCompactor creates a new Compactor instance
func NewCompactor(llmClient llm.Client) *Compactor {
	model := DefaultSummarizationModel
	if val := os.Getenv("COMPACT_MODEL"); val != "" {
		model = val
	}
	return &Compactor{
		llmClient: llmClient,
		model:     model,
	}
}

// EstimateTokens provides a rough token count for messages (4 chars ~= 1 token)
func EstimateTokens(messages []llm.Message) int {
	total := 0
	for _, msg := range messages {
		total += len(msg.Content) / 4
		for _, tc := range msg.ToolCalls {
			total += len(tc.Function.Arguments) / 4
		}
	}
	return total
}

// MaxToolResultLength is the maximum length for tool results before truncation
const MaxToolResultLength = 500

// MaxActiveToolResultLength is the maximum length for tool results in the *active* context window.
// This is larger than the archive limit because we want the agent to see recent results,
// but we still must prevent infinite expansion (e.g. cat huge_file.txt)
const MaxActiveToolResultLength = 2000

// MicroCompact prunes verbose tool call arguments and results while preserving essential information.
// This is a lightweight operation that doesn't require LLM calls.
func MicroCompact(messages []llm.Message) []llm.Message {
	result := make([]llm.Message, 0, len(messages))

	for _, msg := range messages {
		newMsg := llm.Message{
			Role:       msg.Role,
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
		}

		// Prune tool call arguments (keep function name, truncate args)
		if len(msg.ToolCalls) > 0 {
			newMsg.ToolCalls = make([]llm.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				// Keep ID and function name, but truncate long arguments
				args := tc.Function.Arguments
				if len(args) > MaxToolResultLength {
					args = args[:MaxToolResultLength] + "...[truncated]"
				}
				newMsg.ToolCalls[i] = llm.ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: llm.FunctionCallData{
						Name:      tc.Function.Name,
						Arguments: args,
					},
				}
			}
		}

		// Prune tool result messages (truncate long results)
		if msg.Role == "tool" && len(msg.Content) > MaxToolResultLength {
			newMsg.Content = msg.Content[:MaxToolResultLength] + "...[truncated]"
		}

		result = append(result, newMsg)
	}

	return result
}

// Compact summarizes conversation history to reduce token usage
func (c *Compactor) Compact(ctx context.Context, req CompactRequest) (*CompactResponse, error) {
	// Check for cancellation early
	select {
	case <-ctx.Done():
		return &CompactResponse{
			Messages:     req.Messages,
			TokensBefore: EstimateTokens(req.Messages),
			TokensAfter:  EstimateTokens(req.Messages),
		}, ctx.Err()
	default:
	}

	messages := req.Messages

	// Edge case: too few messages to compact
	if len(messages) < 6 {
		return &CompactResponse{
			Messages:     messages,
			TokensBefore: EstimateTokens(messages),
			TokensAfter:  EstimateTokens(messages),
		}, nil
	}

	tokensBefore := EstimateTokens(messages)

	// Step 1: Identify System Prompt and Previous Summary
	var baseSystemPrompt string
	var previousSummary string
	systemMsgIndex := -1

	for i, msg := range messages {
		if msg.Role == "system" {
			systemMsgIndex = i
			// Extract existing summary if present
			summaryStart := strings.Index(msg.Content, "--- CONVERSATION SUMMARY")
			if summaryStart != -1 {
				summaryEnd := strings.Index(msg.Content, "--- END SUMMARY ---")
				if summaryEnd != -1 {
					// Extract the summary content
					headerEnd := strings.Index(msg.Content[summaryStart:], "---\n")
					if headerEnd != -1 {
						previousSummary = strings.TrimSpace(msg.Content[summaryStart+headerEnd+4 : summaryEnd])
					}
					// Base prompt is everything before the summary block
					baseSystemPrompt = strings.TrimSpace(msg.Content[:summaryStart])
				}
			} else {
				baseSystemPrompt = msg.Content
			}
			break
		}
	}

	// Step 2: Identify Retention Cutoff (Keep last N turns active)
	// We want to keep at least the last 1-2 complete turns + current turn
	// A safe heuristic is to keep messages from the 2nd to last User message
	cutoffIndex := -1 // Default: Undefined

	// Scan backwards
	userMsgCount := 0
	lastUserIndex := -1
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role == "user" {
			userMsgCount++
			lastUserIndex = i
			if userMsgCount == 2 { // Keep last 2 user starts (current + previous)
				cutoffIndex = i
				break
			}
		}
	}

	// Fallback: If we didn't find 2 user messages, keep from the last one found
	if cutoffIndex == -1 {
		if lastUserIndex != -1 {
			cutoffIndex = lastUserIndex
		} else {
			// No user messages? Keep everything to be safe
			cutoffIndex = 0
			if systemMsgIndex != -1 {
				cutoffIndex = systemMsgIndex + 1
			}
		}
	}

	// Safety check: Ensure we are actually compacting a meaningful amount of history
	// If cutoffIndex is too close to the system message, we're not splitting enough off to archive.
	if cutoffIndex <= systemMsgIndex+1 {
		// Just return original if we can't meaningfully split
		return &CompactResponse{
			Messages:     messages,
			TokensBefore: tokensBefore,
			TokensAfter:  tokensBefore,
		}, nil
	}

	// Explicitly define what to summarize and what to keep
	// messagesToSummarize: everything AFTER system prompt UP TO cutoff
	var messagesToSummarize []llm.Message
	if systemMsgIndex != -1 {
		if cutoffIndex > systemMsgIndex+1 {
			messagesToSummarize = messages[systemMsgIndex+1 : cutoffIndex]
		}
	} else {
		messagesToSummarize = messages[:cutoffIndex]
	}

	activeMessages := messages[cutoffIndex:]
	// Clone activeMessages to prevent mutation of the input slice (side effect)
	activeMessages = append([]llm.Message(nil), activeMessages...)

	// Safety Pruning for Active Messages
	// Even though these are "active", we cannot allow a single massive tool output to blow up the context.
	// We apply a generous but strict limit to tool outputs in the active window.
	for i, msg := range activeMessages {
		if msg.Role == "tool" && len(msg.Content) > MaxActiveToolResultLength {
			// Create a copy to modify
			newMsg := msg
			newMsg.Content = msg.Content[:MaxActiveToolResultLength] + fmt.Sprintf("...[truncated active tool result: %d chars omitted]", len(msg.Content)-MaxActiveToolResultLength)
			activeMessages[i] = newMsg
		}
	}

	// Step 3: Apply MicroCompact ONLY to the archive messages (before summarization)
	// We don't want to MicroCompact active messages!
	prunedArchive := MicroCompact(messagesToSummarize)

	// Step 4: Build Text for Summarization
	var historyBuilder strings.Builder
	for _, msg := range prunedArchive {
		historyBuilder.WriteString(fmt.Sprintf("[%s]: %s\n", msg.Role, msg.Content))
		// Include tool info for context
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				historyBuilder.WriteString(fmt.Sprintf("(Tool Call: %s)\n", tc.Function.Name))
			}
		}
		if msg.Role == "tool" {
			// For tools, just mention it existed or give brief content if short
			if msg.Name != "" {
				historyBuilder.WriteString(fmt.Sprintf("(Tool Result: %s)\n", msg.Name))
			} else {
				historyBuilder.WriteString("(Tool Result)\n")
			}
		}
	}

	// Step 5: Call LLM
	keepInstruction := ""
	if req.Keep != "" {
		keepInstruction = fmt.Sprintf("\n\nIMPORTANT: Make sure to preserve details about: %s", req.Keep)
	}

	previousSummarySection := ""
	if previousSummary != "" {
		previousSummarySection = fmt.Sprintf(`
Previous Context Summary (incorporate this into your new summary):
%s

`, previousSummary)
	}

	summarizationPrompt := fmt.Sprintf(`Summarize the following conversation history concisely.
This history will be removed from the prompt, so capture ALL critical state, decisions, and code references.

%s

%sRecent Conversation to Archive:
%s

Provide a concise, consolidated summary (2-4 paragraphs). Focus on:
- What has been accomplished
- Function definitions or code snippets active in context
- Errors encountered and resolutions
`, keepInstruction, previousSummarySection, historyBuilder.String())

	llmReq := llm.Request{
		Model: c.model,
		Messages: []llm.Message{
			{Role: "user", Content: summarizationPrompt},
		},
		Temperature: 0.3,
	}

	resp, err := c.llmClient.Chat(ctx, llmReq)
	if err != nil {
		return &CompactResponse{
			Messages:     messages,
			TokensBefore: tokensBefore,
			TokensAfter:  tokensBefore,
		}, fmt.Errorf("summarization failed: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return &CompactResponse{
			Messages:     messages,
			TokensBefore: tokensBefore,
			TokensAfter:  tokensBefore,
		}, fmt.Errorf("empty summarization response")
	}

	newSummary := resp.Choices[0].Message.Content

	// Step 6: Construct Final Messages
	// 1. New System Message (Base + New Summary)
	// 2. Verified Active Messages (Verbatim)

	// Calculate stats
	summaryTokens := len(newSummary) / 4
	recentTokens := EstimateTokens(activeMessages)
	estTokensAfter := summaryTokens + recentTokens // + system prompt overhead (constant-ish)
	savedEst := tokensBefore - estTokensAfter
	if savedEst < 0 {
		savedEst = 0
	}

	compactTime := time.Now()
	compactedSystemContent := fmt.Sprintf("%s\n\n--- CONVERSATION SUMMARY (Saved ~%d tokens | Compacted: %s) ---\n%s\n--- END SUMMARY ---", baseSystemPrompt, savedEst, compactTime.Format("Jan 2, 3:04 PM"), newSummary)

	finalMessages := []llm.Message{
		{Role: "system", Content: compactedSystemContent, Timestamp: &compactTime},
	}
	finalMessages = append(finalMessages, activeMessages...)

	return &CompactResponse{
		Messages:     finalMessages,
		TokensBefore: tokensBefore,
		TokensAfter:  EstimateTokens(finalMessages),
	}, nil
}
