package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/nexuscrm/mcp/pkg/mcp"
)

// NewHandler creates a new HTTP handler for the MCP server
func NewHandler(bus *ToolBusService) http.Handler {
	// 2. Create MCP Server
	server := mcp.NewServer()

	// 3. Register Routes
	server.Register("tools/list", bus.HandleListTools)
	server.Register("tools/call", bus.HandleCallTool)

	// Add other standard routes
	server.Register("ping", func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "pong", nil
	})

	log.Println("MCP Server initialized with ToolBus")
	return server
}
