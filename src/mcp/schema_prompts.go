// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// PromptArgument describes an argument for a prompt.
type PromptArgument struct {
	// Name is the argument identifier.
	Name string `json:"name"`
	// Description is a human-readable explanation of the argument.
	Description string `json:"description,omitempty"`
	// Required indicates whether the argument must be provided.
	Required bool `json:"required,omitempty"`
}

// Prompt describes an available prompt.
type Prompt struct {
	// Name is the unique prompt identifier.
	Name string `json:"name"`
	// Description explains what the prompt does.
	Description string `json:"description,omitempty"`
	// Arguments lists the template parameters accepted by the prompt.
	Arguments []PromptArgument `json:"arguments,omitempty"`
}

// PromptMessage represents a message returned from a prompt.
type PromptMessage struct {
	// Role is the speaker role ("user" or "assistant").
	Role Role `json:"role"`
	// Content holds the message content block.
	Content Content `json:"content"`
}

// PromptReference refers to a prompt.
type PromptReference struct {
	// Type must be "ref/prompt".
	Type string `json:"type"`
	// Name is the referenced prompt name.
	Name string `json:"name"`
}

// GetPromptRequest requests a specific prompt.
type GetPromptRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "prompts/get".
	Method string `json:"method"`
	// Params holds the prompt name and template arguments.
	Params struct {
		// Name is the prompt name to retrieve.
		Name string `json:"name"`
		// Arguments maps argument names to values for template instantiation.
		Arguments map[string]string `json:"arguments,omitempty"`
	} `json:"params"`
}

// GetPromptResult is the response from getting a prompt.
type GetPromptResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Description is an optional description of the rendered prompt.
	Description string `json:"description,omitempty"`
	// Messages lists the rendered conversation messages.
	Messages []PromptMessage `json:"messages"`
}

// ListPromptsRequest requests a paginated list of prompts.
type ListPromptsRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "prompts/list".
	Method string `json:"method"`
	// Params holds optional pagination parameters.
	Params struct {
		// Cursor is the pagination cursor token.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListPromptsResult is the paginated list of prompts.
type ListPromptsResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// NextCursor is the opaque cursor for the next page of prompts.
	NextCursor Cursor `json:"nextCursor,omitempty"`
	// Prompts lists the available prompt definitions.
	Prompts []Prompt `json:"prompts"`
}

// PromptListChangedNotification is sent when the prompt list changes.
type PromptListChangedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/prompts/list_changed".
	Method string `json:"method"`
	// Params holds optional metadata.
	Params struct {
		// Meta contains optional protocol metadata.
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}
