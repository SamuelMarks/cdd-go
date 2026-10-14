// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// Tool describes an available tool.
type Tool struct {
	// Name is the unique identifier of the tool.
	Name string `json:"name"`
	// Description is a human-readable explanation of the tool's purpose.
	Description string `json:"description,omitempty"`
	// InputSchema defines the JSON Schema for the tool's input arguments.
	InputSchema any `json:"inputSchema"`
}

// CallToolRequest represents a request to call a tool.
type CallToolRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "tools/call".
	Method string `json:"method"`
	// Params holds the tool name and invocation arguments.
	Params struct {
		// Name is the name of the tool to invoke.
		Name string `json:"name"`
		// Arguments maps parameter names to argument values.
		Arguments map[string]any `json:"arguments,omitempty"`
	} `json:"params"`
}

// CallToolResult represents the result of a tool call.
type CallToolResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Content lists the content blocks returned by the tool.
	Content []Content `json:"content"`
	// IsError indicates whether the tool execution resulted in an error.
	IsError bool `json:"isError,omitempty"`
}

// ToolListChangedNotification sent when the tool list changes.
type ToolListChangedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/tools/list_changed".
	Method string `json:"method"`
	// Params holds optional metadata.
	Params struct {
		// Meta contains optional protocol metadata.
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ListToolsRequest requests a list of tools.
type ListToolsRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "tools/list".
	Method string `json:"method"`
	// Params holds optional pagination parameters.
	Params struct {
		// Cursor is the pagination cursor token.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListToolsResult returns a paginated list of tools.
type ListToolsResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// NextCursor is the opaque cursor for the next page of tools.
	NextCursor Cursor `json:"nextCursor,omitempty"`
	// Tools lists the available tool definitions.
	Tools []Tool `json:"tools"`
}
