// Package server provides HTTP and SSE server bridging for Model Context Protocol (MCP) messages.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPServer wraps HTTP bridging for MCP.
type MCPServer struct {
	// Handler processes incoming JSON-RPC MCP messages and returns a response message.
	Handler func(req *mcp.Message) (*mcp.Response, error)
}

// NewMCPServer creates a new SSE/HTTP server bridge.
// It returns the initialized MCPServer.
func NewMCPServer() *MCPServer {
	return &MCPServer{}
}

// HandleSSE writes the initial Server-Sent Events endpoint handshake frame.
// Parameter w is the HTTP response writer, and r is the incoming HTTP request.
func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "event: endpoint")
	fmt.Fprintln(w, "data: /mcp/message")
	fmt.Fprintln(w)
}

// HandleMessage handles HTTP POST to the message endpoint, mapping HTTP auth to MCP context.
// Parameter w is the HTTP response writer, and r is the incoming HTTP request.
func (s *MCPServer) HandleMessage(w http.ResponseWriter, r *http.Request) {
	var req mcp.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if s.Handler != nil {
		resp, err := s.Handler(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
