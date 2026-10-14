// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// CompleteRequest is a request to complete an argument.
type CompleteRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "completion/complete".
	Method string `json:"method"`
	// Params holds the reference and argument to complete.
	Params struct {
		// Ref identifies the prompt or resource being completed.
		Ref any `json:"ref"`
		// Argument holds the argument name and partial value.
		Argument struct {
			// Name is the argument name.
			Name string `json:"name"`
			// Value is the current partial argument value.
			Value string `json:"value"`
		} `json:"argument"`
	} `json:"params"`
}

// CompleteResult is the response for a completion request.
type CompleteResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Completion holds the matched completion values and pagination info.
	Completion struct {
		// Values lists the suggested completion strings.
		Values []string `json:"values"`
		// Total is the optional total number of available matches.
		Total int `json:"total,omitempty"`
		// HasMore indicates whether additional completion values exist beyond Values.
		HasMore bool `json:"hasMore,omitempty"`
	} `json:"completion"`
}
