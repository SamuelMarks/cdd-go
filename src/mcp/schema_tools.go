// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import "encoding/json"

// ToolInputSchema defines the strongly-typed JSON Schema for a tool's input arguments.
type ToolInputSchema struct {
	// Type must be "object" for MCP tool input schemas.
	Type string `json:"type"`
	// Properties maps argument names to their JSON Schema definitions.
	Properties map[string]ToolPropertySchema `json:"properties,omitempty"`
	// Required lists the required argument names.
	Required []string `json:"required,omitempty"`
}

// ToolPropertySchema describes a single property within a ToolInputSchema.
type ToolPropertySchema struct {
	// Type is the JSON Schema primitive type (e.g., "string", "integer", "boolean").
	Type string `json:"type"`
	// Description explains the purpose of the property.
	Description string `json:"description,omitempty"`
}

// Tool describes an available tool.
type Tool struct {
	// Name is the unique identifier of the tool.
	Name string `json:"name"`
	// Description is a human-readable explanation of the tool's purpose.
	Description string `json:"description,omitempty"`
	// InputSchema defines the JSON Schema for the tool's input arguments.
	InputSchema ToolInputSchema `json:"inputSchema"`
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
		// Arguments maps parameter names to string argument values.
		Arguments map[string]string `json:"arguments,omitempty"`
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

// UnmarshalJSON deserializes CallToolResult and its polymorphic Content slice.
// Parameter data is the raw JSON bytes, and it returns any unmarshal error.
func (r *CallToolResult) UnmarshalJSON(data []byte) error {
	var raw struct {
		Meta    *Meta             `json:"_meta,omitempty"`
		Content []json.RawMessage `json:"content"`
		IsError bool              `json:"isError,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Meta = raw.Meta
	r.IsError = raw.IsError
	r.Content = make([]Content, 0, len(raw.Content))
	for _, item := range raw.Content {
		c, err := UnmarshalContent(item)
		if err != nil {
			return err
		}
		r.Content = append(r.Content, c)
	}
	return nil
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
