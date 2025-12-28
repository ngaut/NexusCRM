package server

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/nexuscrm/mcp/pkg/mcp"
)

func (s *ToolBusService) handleContextAdd(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	filesRaw, ok := req.Arguments["files"].([]interface{})
	if !ok {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "files array is required"}}}, nil
	}

	session := s.contextStore.GetSession(token)
	var added []string
	var errors []string

	for _, f := range filesRaw {
		path, ok := f.(string)
		if !ok {
			continue
		}
		if err := session.AddFile(path); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to add %s: %v", path, err))
		} else {
			added = append(added, path)
		}
	}

	msg := fmt.Sprintf("Added %d files to context.", len(added))
	if len(errors) > 0 {
		msg += fmt.Sprintf("\nErrors:\n%s", strings.Join(errors, "\n"))
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: msg}},
	}, nil
}

func (s *ToolBusService) handleContextRemove(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	filesRaw, ok := req.Arguments["files"].([]interface{})
	if !ok {
		return mcp.CallToolResult{IsError: true, Content: []mcp.Content{{Type: "text", Text: "files array is required"}}}, nil
	}

	session := s.contextStore.GetSession(token)
	count := 0
	for _, f := range filesRaw {
		path, ok := f.(string)
		if !ok {
			continue
		}
		session.RemoveFile(path)
		count++
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Removed %d files from context.", count)}},
	}, nil
}

func (s *ToolBusService) handleContextList(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	session := s.contextStore.GetSession(token)
	items := session.ListItems()
	totalTokens := session.GetTotalTokens()

	// Sort by path
	sort.Slice(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Active Context (%d files, ~%d tokens):\n", len(items), totalTokens))
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s (~%d tokens)\n", item.Path, item.TokenSize))
	}

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: sb.String()}},
	}, nil
}

func (s *ToolBusService) handleContextClear(ctx context.Context, req mcp.CallToolParams) (mcp.CallToolResult, error) {
	token, err := s.getAuthToken(ctx)
	if err != nil {
		return mcp.CallToolResult{}, err
	}

	session := s.contextStore.GetSession(token)
	session.Clear()

	return mcp.CallToolResult{
		Content: []mcp.Content{{Type: "text", Text: "Context cleared."}},
	}, nil
}
