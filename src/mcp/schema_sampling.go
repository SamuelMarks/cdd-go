// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// ModelHint provides a hint for model selection.
type ModelHint struct {
	// Name is the suggested model name or substring (e.g., "claude-3-5-sonnet").
	Name string `json:"name,omitempty"`
}

// ModelPreferences provides preferences for model selection.
type ModelPreferences struct {
	// Hints is an ordered list of model hints.
	Hints []ModelHint `json:"hints,omitempty"`
	// CostPriority indicates how much to prioritize low cost (0.0 to 1.0).
	CostPriority float64 `json:"costPriority,omitempty"`
	// SpeedPriority indicates how much to prioritize low latency (0.0 to 1.0).
	SpeedPriority float64 `json:"speedPriority,omitempty"`
	// IntelligencePriority indicates how much to prioritize model capability (0.0 to 1.0).
	IntelligencePriority float64 `json:"intelligencePriority,omitempty"`
}

// CreateMessageRequestParams holds the parameters for sampling.
type CreateMessageRequestParams struct {
	// Messages is the conversation history to sample from.
	Messages []SamplingMessage `json:"messages"`
	// ModelPreferences specifies optional model selection preferences.
	ModelPreferences *ModelPreferences `json:"modelPreferences,omitempty"`
	// SystemPrompt is the optional system prompt to apply.
	SystemPrompt string `json:"systemPrompt,omitempty"`
	// IncludeContext specifies which MCP context to include ("none", "thisServer", "allServers").
	IncludeContext string `json:"includeContext,omitempty"`
	// Temperature controls sampling randomness.
	Temperature float64 `json:"temperature,omitempty"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"maxTokens"`
	// StopSequences lists sequences that halt token generation.
	StopSequences []string `json:"stopSequences,omitempty"`
	// Metadata holds provider-specific sampling metadata.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// FullCreateMessageRequest represents a full request to sample from an LLM.
type FullCreateMessageRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "sampling/createMessage".
	Method string `json:"method"`
	// Params holds the full sampling request parameters.
	Params CreateMessageRequestParams `json:"params"`
}

// CreateMessageResult is the response from sampling.
type CreateMessageResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Role is the role of the generated message ("assistant").
	Role Role `json:"role"`
	// Content holds the generated message content block.
	Content Content `json:"content"`
	// Model is the identifier of the model that generated the response.
	Model string `json:"model"`
	// StopReason explains why generation stopped (e.g., "endTurn", "maxTokens").
	StopReason string `json:"stopReason,omitempty"`
}
