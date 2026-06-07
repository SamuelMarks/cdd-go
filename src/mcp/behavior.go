package mcp

// ProgressNotification represents a progress event.
type ProgressNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		ProgressToken interface{} `json:"progressToken"`
		Progress      float64     `json:"progress"`
		Total         float64     `json:"total,omitempty"`
	} `json:"params"`
}

// SecurityManager defines boundaries and approvals.
type SecurityManager struct {
	AllowedRoots    []string
	RequireApproval bool
}

// IsPathAllowed checks if a path is within the allowed roots.
func (sm *SecurityManager) IsPathAllowed(path string) bool {
	if len(sm.AllowedRoots) == 0 {
		return false // Secure by default
	}
	for _, root := range sm.AllowedRoots {
		if len(path) >= len(root) && path[:len(root)] == root {
			return true
		}
	}
	return false
}

// RequiresToolApproval checks if a tool call needs human approval.
func (sm *SecurityManager) RequiresToolApproval(toolName string) bool {
	return sm.RequireApproval
}

// Cursor is used for pagination.
type Cursor string

// PaginatedRequest provides a standard struct for requests that handle cursors.
type PaginatedRequest struct {
	Method string `json:"method"`
	Params struct {
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// PaginatedResult provides a standard struct for handling cursors.
type PaginatedResult struct {
	Meta       *Meta       `json:"_meta,omitempty"`
	NextCursor Cursor      `json:"nextCursor,omitempty"`
	Items      interface{} `json:"items"`
}

// SamplingMessage represents a message for human-in-the-loop sampling.
type SamplingMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CreateMessageRequest represents a request to sample from an LLM.
type CreateMessageRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  struct {
		Messages     []SamplingMessage `json:"messages"`
		SystemPrompt string            `json:"systemPrompt,omitempty"`
		MaxTokens    int               `json:"maxTokens,omitempty"`
	} `json:"params"`
}

// ProtocolHandler handles URI protocols.
type ProtocolHandler struct {
	SupportedSchemes []string
}

// IsSchemeSupported checks if a URI scheme is supported.
func (ph *ProtocolHandler) IsSchemeSupported(uri string) bool {
	for _, scheme := range ph.SupportedSchemes {
		if len(uri) >= len(scheme) && uri[:len(scheme)] == scheme {
			return true
		}
	}
	return false
}
