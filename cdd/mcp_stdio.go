// Package cdd provides the core orchestration for OpenAPI and Go AST code generation.
package cdd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPRequest is an alias for mcp.Request representing an incoming MCP tool call.
type MCPRequest = mcp.Request

// MCPResponse is an alias for mcp.Response representing an outgoing MCP response.
type MCPResponse = mcp.Response

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

// mcpStdin is an injectable reader for ServeMCPStdio.
var mcpStdin io.Reader = os.Stdin

// mcpStdout is an injectable writer for ServeMCPStdio.
var mcpStdout io.Writer = os.Stdout

// mcpStderr is an injectable error writer for ServeMCPStdio.
var mcpStderr io.Writer = os.Stderr

// ServeMCPStdio serves the MCP tool interface via standard I/O.
// Parameter args is the slice of CLI arguments, and it returns any execution error.
func ServeMCPStdio(args []string) error {
	return ServeMCP(mcpStdin, mcpStdout, mcpStderr)
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

		var resultRaw json.RawMessage
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
			initRes.ServerInfo.Version = "0.0.4"
			resultRaw, _ = json.Marshal(initRes)
		case "tools/list":
			inputSchema := mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]mcp.ToolPropertySchema{
					"input":  {Type: "string"},
					"output": {Type: "string"},
				},
				Required: []string{"input", "output"},
			}
			listRes := mcp.ListToolsResult{
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
					{
						Name:        "cdd_generate_arazzo_sdk",
						Description: "Generate a workflow SDK from an Arazzo 1.1.0 spec",
						InputSchema: inputSchema,
					},
					{
						Name:        "cdd_sync_arazzo",
						Description: "Extract an Arazzo 1.1.0 spec from Go source code",
						InputSchema: inputSchema,
					},
				},
			}
			resultRaw, _ = json.Marshal(listRes)
		case "tools/call":
			var params MCPCallToolParams
			if err := json.Unmarshal(req.Params, &params); err == nil {
				switch params.Name {
				case "cdd_generate_sdk":
					if err := FromOpenAPI("to_sdk", params.Arguments.Input, params.Arguments.Output, false, false, false, false); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						resultRaw, _ = json.Marshal(mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "SDK successfully generated"},
							},
						})
					}
				case "cdd_sync_schema":
					if err := ToOpenAPI(params.Arguments.Input, params.Arguments.Output); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						resultRaw, _ = json.Marshal(mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "OpenAPI schema successfully extracted"},
							},
						})
					}
				case "cdd_generate_arazzo_sdk":
					if err := FromArazzo("to_sdk", params.Arguments.Input, params.Arguments.Output, false, false, false); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						resultRaw, _ = json.Marshal(mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "Arazzo SDK successfully generated"},
							},
						})
					}
				case "cdd_sync_arazzo":
					if err := ToArazzo(params.Arguments.Input, params.Arguments.Output); err != nil {
						errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
					} else {
						resultRaw, _ = json.Marshal(mcp.CallToolResult{
							Content: []mcp.Content{
								mcp.TextContent{Type: "text", Text: "Arazzo specification successfully extracted"},
							},
						})
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
			Result:  resultRaw,
			Error:   errResp,
		}

		outBytes, _ := json.Marshal(resp)
		fmt.Fprintln(out, string(outBytes))
	}

	return nil
}
