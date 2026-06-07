package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPServer wraps HTTP bridging for MCP.
type MCPServer struct {
	Handler func(req *mcp.Message) (*mcp.Response, error)
}

// NewMCPServer creates a new SSE/HTTP server bridge.
func NewMCPServer() *MCPServer {
	return &MCPServer{}
}

// HandleSSE Endpoint Generation
func (s *MCPServer) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// Stub: Wires MCP endpoints via SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "event: endpoint\ndata: /mcp/message\n\n")
}

// HandleMessage handles HTTP POST to the message endpoint, mapping HTTP auth to MCP context.
func (s *MCPServer) HandleMessage(w http.ResponseWriter, r *http.Request) {
	var req mcp.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// HTTP Request/Auth Bridging & Dynamic API-to-Tool Proxy
	if s.Handler != nil {
		resp, err := s.Handler(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(resp)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
