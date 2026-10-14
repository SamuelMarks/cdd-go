// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// ProgressNotification represents a progress event.
type ProgressNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is the notification method ("$/progress" or "notifications/progress").
	Method string `json:"method"`
	// Params holds the progress token and completion metrics.
	Params struct {
		// ProgressToken identifies the operation whose progress is being reported.
		ProgressToken ProgressToken `json:"progressToken"`
		// Progress is the current progress value.
		Progress float64 `json:"progress"`
		// Total is the optional total progress target.
		Total float64 `json:"total,omitempty"`
	} `json:"params"`
}

// SecurityManager defines boundaries and approvals.
type SecurityManager struct {
	// AllowedRoots lists the permitted root filesystem/URI prefixes.
	AllowedRoots []string
	// RequireApproval specifies whether tool execution requires human approval.
	RequireApproval bool
}

// IsPathAllowed checks if a path is within the allowed roots.
// Parameter path is the target path or URI, and it returns true if permitted.
func (sm *SecurityManager) IsPathAllowed(path string) bool {
	if len(sm.AllowedRoots) == 0 {
		return false
	}
	for _, root := range sm.AllowedRoots {
		if len(path) >= len(root) && path[:len(root)] == root {
			return true
		}
	}
	return false
}

// RequiresToolApproval checks if a tool call needs human approval.
// Parameter toolName is the tool name, and it returns true if approval is required.
func (sm *SecurityManager) RequiresToolApproval(toolName string) bool {
	return sm.RequireApproval
}

// Cursor is used for pagination.
type Cursor string

// PaginatedRequest provides a standard struct for requests that handle cursors.
type PaginatedRequest struct {
	// Method is the paginated RPC method name.
	Method string `json:"method"`
	// Params holds the pagination cursor.
	Params struct {
		// Cursor is the opaque pagination cursor token.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// PaginatedResult provides a standard struct for handling cursors.
type PaginatedResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// NextCursor is the opaque cursor for the next page of results.
	NextCursor Cursor `json:"nextCursor,omitempty"`
	// Items holds the paginated collection items.
	Items any `json:"items"`
}

// SamplingMessage represents a message for human-in-the-loop sampling.
type SamplingMessage struct {
	// Role is the conversational role ("user" or "assistant").
	Role string `json:"role"`
	// Content is the textual content of the message.
	Content string `json:"content"`
}

// CreateMessageRequest represents a request to sample from an LLM.
type CreateMessageRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method is the sampling method ("sampling/createMessage").
	Method string `json:"method"`
	// Params holds the sampling parameters.
	Params struct {
		// Messages is the conversation history to sample from.
		Messages []SamplingMessage `json:"messages"`
		// SystemPrompt is the optional system prompt.
		SystemPrompt string `json:"systemPrompt,omitempty"`
		// MaxTokens is the maximum number of tokens to generate.
		MaxTokens int `json:"maxTokens,omitempty"`
	} `json:"params"`
}

// ProtocolHandler handles URI protocols.
type ProtocolHandler struct {
	// SupportedSchemes lists the URI schemes supported by the handler (e.g., "file://", "https://").
	SupportedSchemes []string
}

// IsSchemeSupported checks if a URI scheme is supported.
// Parameter uri is the target URI string, and it returns true if its scheme is supported.
func (ph *ProtocolHandler) IsSchemeSupported(uri string) bool {
	for _, scheme := range ph.SupportedSchemes {
		if len(uri) >= len(scheme) && uri[:len(scheme)] == scheme {
			return true
		}
	}
	return false
}
