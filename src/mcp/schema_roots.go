// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// Root describes a single root object.
type Root struct {
	// URI is the root directory or resource URI (typically starting with "file://").
	URI string `json:"uri"`
	// Name is an optional human-readable label for the root.
	Name string `json:"name,omitempty"`
}

// ListRootsRequest requests a list of roots.
type ListRootsRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "roots/list".
	Method string `json:"method"`
	// Params holds optional metadata.
	Params struct {
		// Meta contains optional protocol metadata.
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ListRootsResult returns a list of roots.
type ListRootsResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Roots lists the workspace root objects.
	Roots []Root `json:"roots"`
}

// RootsListChangedNotification notifies the server that the roots list has changed.
type RootsListChangedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/roots/list_changed".
	Method string `json:"method"`
	// Params holds optional metadata.
	Params struct {
		// Meta contains optional protocol metadata.
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}
