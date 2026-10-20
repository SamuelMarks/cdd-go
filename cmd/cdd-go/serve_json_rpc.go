// Package main provides the CLI entrypoint and subcommand dispatcher for cdd-go.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/SamuelMarks/cdd-go/cdd"
	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// RPCRequest is a strongly-typed JSON-RPC 2.0 request alias over mcp.Request.
type RPCRequest = mcp.Request

// RPCResponse is a strongly-typed JSON-RPC 2.0 response alias over mcp.Response.
type RPCResponse = mcp.Response

// RPCFromOpenAPIParams represents map-structured parameters for the from_openapi RPC method.
type RPCFromOpenAPIParams struct {
	// Subcommand specifies the generation target (e.g., "to_sdk", "to_server", "to_sdk_cli").
	Subcommand string `json:"subcommand"`
	// Input is the path to the input OpenAPI specification file.
	Input string `json:"input"`
	// InputDir is the directory containing OpenAPI specifications.
	InputDir string `json:"input_dir"`
	// Output is the output directory path.
	Output string `json:"output"`
	// NoGithubActions disables generating GitHub Actions workflow scaffolding.
	NoGithubActions bool `json:"no_github_actions"`
	// NoInstallablePackage disables generating go.mod scaffolding.
	NoInstallablePackage bool `json:"no_installable_package"`
	// Tests enables generating test suites and mocks.
	Tests bool `json:"tests"`
}

// httpListenAndServe is an injectable function variable wrapping http.ListenAndServe.
var httpListenAndServe = http.ListenAndServe

// newJSONRPCHandler constructs a dedicated http.Handler for serving JSON-RPC 2.0 requests without mutating http.DefaultServeMux or os.Stdout.
// It returns the configured http.Handler.
func newJSONRPCHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var result json.RawMessage
		var errResp *mcp.Error

		switch req.Method {
		case "version":
			result = json.RawMessage(`"0.0.4"`)
		case "to_docs_json":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				var buf bytes.Buffer
				err := runToDocsJSONWriter(params, &buf)
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else if json.Valid(buf.Bytes()) && len(bytes.TrimSpace(buf.Bytes())) > 0 {
					result = json.RawMessage(bytes.TrimSpace(buf.Bytes()))
				} else {
					b, _ := json.Marshal(buf.String())
					result = json.RawMessage(b)
				}
			} else {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params for to_docs_json"}
			}
		case "to_openapi":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"to_openapi"}, params...))
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else {
					result = json.RawMessage(`"Success"`)
				}
			} else {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params for to_openapi"}
			}
		case "to_arazzo":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"to_arazzo"}, params...))
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else {
					result = json.RawMessage(`"Success"`)
				}
			} else {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params for to_arazzo"}
			}
		case "from_arazzo":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"from_arazzo"}, params...))
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else {
					result = json.RawMessage(`"Success"`)
				}
			} else {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params for from_arazzo"}
			}
		case "from_openapi":
			var params []string
			var mapParams RPCFromOpenAPIParams
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"from_openapi"}, params...))
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else {
					result = json.RawMessage(`"Success"`)
				}
			} else if err := json.Unmarshal(req.Params, &mapParams); err == nil {
				var args []string
				args = append(args, "from_openapi")
				if mapParams.Subcommand != "" {
					args = append(args, mapParams.Subcommand)
				}
				if mapParams.Input != "" {
					args = append(args, "-i", mapParams.Input)
				}
				if mapParams.InputDir != "" {
					args = append(args, "--input-dir", mapParams.InputDir)
				}
				if mapParams.Output != "" {
					args = append(args, "-o", mapParams.Output)
				}
				if mapParams.NoGithubActions {
					args = append(args, "--no-github-actions")
				}
				if mapParams.NoInstallablePackage {
					args = append(args, "--no-installable-package")
				}
				if mapParams.Tests {
					args = append(args, "--tests")
				}
				err := run(args)
				if err != nil {
					errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
				} else {
					result = json.RawMessage(`"Success"`)
				}
			} else {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params for from_openapi"}
			}
		case "cdd_generate_sdk", "cdd_sync_schema", "cdd_generate_arazzo_sdk", "cdd_sync_arazzo":
			var toolParams cdd.MCPToolParams
			if err := json.Unmarshal(req.Params, &toolParams); err != nil {
				errResp = &mcp.Error{Code: mcp.ErrInvalidParams, Message: "Invalid params"}
			} else if msg, err := cdd.ExecuteGeneratorRouter(req.Method, toolParams); err != nil {
				errResp = &mcp.Error{Code: mcp.ErrInternal, Message: err.Error()}
			} else {
				b, _ := json.Marshal(msg)
				result = json.RawMessage(b)
			}
		default:
			errResp = &mcp.Error{Code: mcp.ErrMethodNotFound, Message: "Method not found"}
		}

		resp := RPCResponse{
			JSONRPC: "2.0",
			Result:  result,
			Error:   errResp,
			ID:      req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	return mux
}

// runServeJSONRPC parses CLI flags and starts the JSON-RPC HTTP server.
// Parameter args is the slice of CLI arguments, and it returns any server error.
func runServeJSONRPC(args []string) error {
	fs := flag.NewFlagSet("serve_json_rpc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var port int
	var listen string

	fs.IntVar(&port, "port", envOrDefaultInt("CDD_PORT", 8080), "Port to listen on")
	fs.IntVar(&port, "p", envOrDefaultInt("CDD_PORT", 8080), "Port to listen on (shorthand)")
	fs.StringVar(&listen, "listen", envOrDefault("CDD_LISTEN", "127.0.0.1"), "Address to listen on")
	fs.StringVar(&listen, "l", envOrDefault("CDD_LISTEN", "127.0.0.1"), "Address to listen on (shorthand)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	addr := fmt.Sprintf("%s:%d", listen, port)
	fmt.Println("Starting JSON-RPC server on", addr)

	return httpListenAndServe(addr, newJSONRPCHandler())
}
