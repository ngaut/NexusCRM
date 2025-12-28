package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// HandlerFunc matches the signature of an MCP method handler
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// Server is a generic MCP server
type Server struct {
	handlers map[string]HandlerFunc
	mu       sync.RWMutex
}

// NewServer creates a new MCP Server
func NewServer() *Server {
	return &Server{
		handlers: make(map[string]HandlerFunc),
	}
}

// Register registers a handler for a specific tool/method
func (s *Server) Register(method string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// HandleHTTP manages MCP over HTTP (POST for RPC, GET for SSE handshake)
// Note: For simplicity in Phase 1, we implement standard POST JSON-RPC.
// Full SSE support requires a more complex connection manager.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeError(w, nil, ErrParse, "Parse error")
		return
	}
	defer r.Body.Close()

	var req Request
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeError(w, nil, ErrParse, "Parse error")
		return
	}

	// Route the request
	resp := s.handleRequest(r.Context(), req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleRequest(ctx context.Context, req Request) Response {
	s.mu.RLock()
	handler, ok := s.handlers[req.Method]
	s.mu.RUnlock()

	if !ok {
		return Response{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Error:   &Error{Code: ErrMethodNotFound, Message: "Method not found"},
		}
	}

	result, err := handler(ctx, req.Params)
	if err != nil {
		// If the error is already a standardized MCP error, use it
		if mcpErr, ok := err.(*Error); ok {
			return Response{
				JSONRPC: JSONRPCVersion,
				ID:      req.ID,
				Error:   mcpErr,
			}
		}
		// Otherwise generic internal error
		log.Printf("MCP Internal Error: %v", err)
		return Response{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Error:   &Error{Code: ErrInternal, Message: fmt.Sprintf("Internal error: %v", err)},
		}
	}

	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
	}
}

func (s *Server) writeError(w http.ResponseWriter, id interface{}, code int, msg string) {
	resp := Response{
		JSONRPC: JSONRPCVersion,
		ID:      id,
		Error:   &Error{Code: code, Message: msg},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC errors are 200 OK HTTP
	json.NewEncoder(w).Encode(resp)
}
