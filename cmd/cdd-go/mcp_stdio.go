package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// MCPRequest represents an incoming MCP tool call.
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      interface{}     `json:"id"`
	Params  json.RawMessage `json:"params"`
}

// MCPResponse represents an outgoing MCP response.
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func runServeMCPStdio(args []string) error {
	fmt.Fprintf(os.Stderr, "Starting CDD Generator MCP Server on stdio...\n")
	decoder := json.NewDecoder(os.Stdin)

	for {
		var req MCPRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error decoding request: %v\n", err)
			// Reset decoder on error to avoid infinite loop
			decoder = json.NewDecoder(os.Stdin)
			// Need to consume bad input, but since stdin is blocking, we just break
			break
		}

		var result interface{}
		var errResp interface{}

		if req.Method == "initialize" {
			result = map[string]interface{}{
				"protocolVersion": "1.0.0",
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    "cdd-generator-mcp",
					"version": "0.0.3",
				},
			}
		} else if req.Method == "tools/list" {
			result = map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "cdd_generate_sdk",
						"description": "Generate an SDK from an OpenAPI spec",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"input":  map[string]interface{}{"type": "string"},
								"output": map[string]interface{}{"type": "string"},
							},
							"required": []string{"input", "output"},
						},
					},
					{
						"name":        "cdd_sync_schema",
						"description": "Extract an OpenAPI spec from source code",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"input":  map[string]interface{}{"type": "string"},
								"output": map[string]interface{}{"type": "string"},
							},
							"required": []string{"input", "output"},
						},
					},
				},
			}
		} else if req.Method == "tools/call" {
			var params struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &params); err == nil {
				if params.Name == "cdd_generate_sdk" {
					in, _ := params.Arguments["input"].(string)
					out, _ := params.Arguments["output"].(string)
					err := run([]string{"from_openapi", "to_sdk", "-i", in, "-o", out})
					if err != nil {
						errResp = map[string]interface{}{"code": -32603, "message": err.Error()}
					} else {
						result = map[string]interface{}{
							"content": []map[string]interface{}{
								{"type": "text", "text": "SDK successfully generated"},
							},
						}
					}
				} else if params.Name == "cdd_sync_schema" {
					in, _ := params.Arguments["input"].(string)
					out, _ := params.Arguments["output"].(string)
					err := run([]string{"to_openapi", "-i", in, "-o", out})
					if err != nil {
						errResp = map[string]interface{}{"code": -32603, "message": err.Error()}
					} else {
						result = map[string]interface{}{
							"content": []map[string]interface{}{
								{"type": "text", "text": "OpenAPI schema successfully extracted"},
							},
						}
					}
				} else {
					errResp = map[string]interface{}{"code": -32601, "message": "Method not found"}
				}
			} else {
				errResp = map[string]interface{}{"code": -32602, "message": "Invalid params"}
			}
		} else {
			errResp = map[string]interface{}{"code": -32601, "message": "Method not found"}
		}

		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
			Error:   errResp,
		}

		out, _ := json.Marshal(resp)
		fmt.Printf("%s\n", out)
	}

	return nil
}
