//go:build !js && !wasip1 && !wasm

// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// cddHTTPListenAndServe is an injectable function variable wrapping http.ListenAndServe for ServeJSONRPC.
var cddHTTPListenAndServe = http.ListenAndServe

// NewJSONRPCHandler constructs an http.Handler serving JSON-RPC 2.0 requests for version and tool execution.
// It returns the configured http.Handler.
func NewJSONRPCHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		resp := MCPResponse{JSONRPC: "2.0", ID: req.ID}
		switch req.Method {
		case "version":
			resp.Result = json.RawMessage(`"0.0.4"`)
		default:
			var params MCPToolParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				resp.Error = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params"}
			} else if msg, err := ExecuteGeneratorRouter(req.Method, params); err != nil {
				resp.Error = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
			} else {
				resp.Result, _ = json.Marshal(msg)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	return mux
}

// ServeJSONRPC exposes the CLI interface as a JSON-RPC HTTP server.
// Parameter port is the TCP port number, listen is the bind address, and it returns any server error.
func ServeJSONRPC(port int, listen string) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port: %d", port)
	}
	if strings.TrimSpace(listen) == "" {
		return fmt.Errorf("listen address is required")
	}
	addr := fmt.Sprintf("%s:%d", listen, port)
	return cddHTTPListenAndServe(addr, NewJSONRPCHandler())
}
