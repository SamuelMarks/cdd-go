package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

// RPCRequest represents a JSON-RPC request.
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	ID     interface{}     `json:"id"`
}

// RPCResponse represents a JSON-RPC response.
type RPCResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
	ID     interface{} `json:"id"`
}

func runServerJSONRPC(args []string) error {
	fs := flag.NewFlagSet("server_json_rpc", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var port int
	var listen string

	fs.IntVar(&port, "port", 8080, "Port to listen on")
	fs.StringVar(&listen, "listen", "0.0.0.0", "Address to listen on")

	if err := fs.Parse(args); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", listen, port)
	fmt.Printf("Starting JSON-RPC server on %s\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var result interface{}
		var errResp interface{}

		switch req.Method {
		case "version":
			result = "0.0.1"
		case "to_docs_json":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				oldStdout := os.Stdout
				rPipe, wPipe, _ := os.Pipe()
				os.Stdout = wPipe

				err := runToDocsJSON(params)

				wPipe.Close()
				os.Stdout = oldStdout

				var buf bytes.Buffer
				buf.ReadFrom(rPipe)

				if err != nil {
					errResp = err.Error()
				} else {
					var jsonOut interface{}
					if jErr := json.Unmarshal(buf.Bytes(), &jsonOut); jErr == nil {
						result = jsonOut
					} else {
						result = buf.String()
					}
				}
			} else {
				errResp = "Invalid params for to_docs_json"
			}
		case "to_openapi":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"to_openapi"}, params...))
				if err != nil {
					errResp = err.Error()
				} else {
					result = "Success"
				}
			} else {
				errResp = "Invalid params for to_openapi"
			}
		case "from_openapi":
			var params []string
			if err := json.Unmarshal(req.Params, &params); err == nil {
				err := run(append([]string{"from_openapi"}, params...))
				if err != nil {
					errResp = err.Error()
				} else {
					result = "Success"
				}
			} else {
				errResp = "Invalid params for from_openapi"
			}
		default:
			errResp = "Method not found"
		}

		resp := RPCResponse{
			Result: result,
			Error:  errResp,
			ID:     req.ID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	return http.ListenAndServe(addr, nil)
}
