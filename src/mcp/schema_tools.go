package mcp

// Tool describes an available tool.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"inputSchema"` // JSON schema
}

// CallToolRequest represents a request to call a tool.
type CallToolRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "tools/call"
	Params  struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments,omitempty"`
	} `json:"params"`
}

// CallToolResult represents the result of a tool call.
type CallToolResult struct {
	Meta    *Meta     `json:"_meta,omitempty"`
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// ToolListChangedNotification sent when the tool list changes.
type ToolListChangedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/tools/list_changed"
	Params  struct {
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ListToolsRequest requests a list of tools.
type ListToolsRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "tools/list"
	Params  struct {
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListToolsResult returns a paginated list of tools.
type ListToolsResult struct {
	Meta       *Meta  `json:"_meta,omitempty"`
	NextCursor Cursor `json:"nextCursor,omitempty"`
	Tools      []Tool `json:"tools"`
}
