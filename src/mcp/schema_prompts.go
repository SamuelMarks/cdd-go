package mcp

// PromptArgument describes an argument for a prompt.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// Prompt describes an available prompt.
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptMessage represents a message returned from a prompt.
type PromptMessage struct {
	Role    Role    `json:"role"`
	Content Content `json:"content"`
}

// PromptReference refers to a prompt.
type PromptReference struct {
	Type string `json:"type"` // MUST be "ref/prompt"
	Name string `json:"name"`
}

// GetPromptRequest requests a specific prompt.
type GetPromptRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "prompts/get"
	Params  struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments,omitempty"`
	} `json:"params"`
}

// GetPromptResult is the response from getting a prompt.
type GetPromptResult struct {
	Meta        *Meta           `json:"_meta,omitempty"`
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// ListPromptsRequest requests a paginated list of prompts.
type ListPromptsRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "prompts/list"
	Params  struct {
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListPromptsResult is the paginated list of prompts.
type ListPromptsResult struct {
	Meta       *Meta    `json:"_meta,omitempty"`
	NextCursor Cursor   `json:"nextCursor,omitempty"`
	Prompts    []Prompt `json:"prompts"`
}

// PromptListChangedNotification is sent when the prompt list changes.
type PromptListChangedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/prompts/list_changed"
	Params  struct {
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}
