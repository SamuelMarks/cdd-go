// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPRequest represents an incoming MCP tool call.
type MCPRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is the MCP method name (e.g., "initialize", "tools/list", "tools/call").
	Method string `json:"method"`
	// ID is the unique request identifier.
	ID mcp.RequestID `json:"id"`
	// Params holds the raw JSON parameters.
	Params json.RawMessage `json:"params"`
}

// MCPResponse represents an outgoing MCP response.
type MCPResponse struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the response identifier matching the request.
	ID mcp.RequestID `json:"id"`
	// Result holds the successful execution payload.
	Result any `json:"result,omitempty"`
	// Error holds the JSON-RPC error object if the call failed.
	Error *mcp.Error `json:"error,omitempty"`
}

// MCPToolParams represents the arguments passed to cdd_generate_sdk or cdd_sync_schema.
type MCPToolParams struct {
	// Input is the input file or directory path.
	Input string `json:"input"`
	// Output is the output file or directory path.
	Output string `json:"output"`
}

// MCPCallToolParams represents the parameters for the tools/call MCP method.
type MCPCallToolParams struct {
	// Name is the tool name to invoke.
	Name string `json:"name"`
	// Arguments holds the strongly-typed tool arguments.
	Arguments MCPToolParams `json:"arguments"`
}

// ServeMCPStdio serves the MCP tool interface via standard I/O.
// Parameter args is the slice of CLI arguments, and it returns any execution error.
func ServeMCPStdio(args []string) error {
	return ServeMCP(os.Stdin, os.Stdout, os.Stderr)
}

// ServeMCP serves the MCP tool interface over the provided reader, writer, and error writer.
// Parameter in is the input stream, out is the output stream, and errW is the diagnostic stream.
func ServeMCP(in io.Reader, out io.Writer, errW io.Writer) error {
	fmt.Fprintln(errW, "Starting CDD Generator MCP Server on stdio...")
	decoder := json.NewDecoder(in)

	for {
		var req MCPRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintln(errW, "Error decoding request:", err)
			break
		}

		var result any
		var errResp *mcp.Error

		switch req.Method {
		case "initialize":
			var initRes mcp.InitializeResult
			initRes.ProtocolVersion = "1.0.0"
			initRes.Capabilities = mcp.ServerCapabilities{
				Tools: &struct {
					ListChanged bool `json:"listChanged,omitempty"`
				}{},
			}
			initRes.ServerInfo.Name = "cdd-generator-mcp"
			initRes.ServerInfo.Version = "0.0.3"
			result = initRes
		case "tools/list":
			inputSchema := map[string]any{
				"type": "object",
				"properties": map[string]any{
					"input":  map[string]string{"type": "string"},
					"output": map[string]string{"type": "string"},
				},
				"required": []string{"input", "output"},
			}
			result = mcp.ListToolsResult{
				Tools: []mcp.Tool{
					{
						Name:        "cdd_generate_sdk",
						Description: "Generate an SDK from an OpenAPI spec",
						InputSchema: inputSchema,
					},
					{
						Name:        "cdd_sync_schema",
						Description: "Extract an OpenAPI spec from source code",
						InputSchema: inputSchema,
					},
				},
			}
		case "tools/call":
			var params MCPCallToolParams
			if err := json.Unmarshal(req.Params, &params); err == nil {
				switch params.Name {
				case "cdd_generate_sdk":
					if err := FromOpenAPI("to_sdk", params.Arguments.Input, params.Arguments.Output, false, false, false, false); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						result = mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "SDK successfully generated"},
							},
						}
					}
				case "cdd_sync_schema":
					if err := ToOpenAPI(params.Arguments.Input, params.Arguments.Output); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						result = mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "OpenAPI schema successfully extracted"},
							},
						}
					}
				default:
					errResp = mcp.MapErrorCode(mcp.ErrMethodNotFound, "")
				}
			} else {
				errResp = mcp.MapErrorCode(mcp.ErrInvalidParams, "")
			}
		default:
			errResp = mcp.MapErrorCode(mcp.ErrMethodNotFound, "")
		}

		resp := MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
			Error:   errResp,
		}

		outBytes, _ := json.Marshal(resp)
		fmt.Fprintln(out, string(outBytes))
	}

	return nil
}
