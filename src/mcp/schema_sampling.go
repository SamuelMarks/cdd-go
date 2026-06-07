package mcp

// ModelHint provides a hint for model selection.
type ModelHint struct {
	Name string `json:"name,omitempty"`
}

// ModelPreferences provides preferences for model selection.
type ModelPreferences struct {
	Hints                []ModelHint `json:"hints,omitempty"`
	CostPriority         float64     `json:"costPriority,omitempty"`
	SpeedPriority        float64     `json:"speedPriority,omitempty"`
	IntelligencePriority float64     `json:"intelligencePriority,omitempty"`
}

// CreateMessageRequestParams holds the parameters for sampling.
type CreateMessageRequestParams struct {
	Messages         []SamplingMessage      `json:"messages"`
	ModelPreferences *ModelPreferences      `json:"modelPreferences,omitempty"`
	SystemPrompt     string                 `json:"systemPrompt,omitempty"`
	IncludeContext   string                 `json:"includeContext,omitempty"` // "none" or "thisServer"
	Temperature      float64                `json:"temperature,omitempty"`
	MaxTokens        int                    `json:"maxTokens"`
	StopSequences    []string               `json:"stopSequences,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CreateMessageRequest represents a request to sample from an LLM.
// Note: Overwrites the simple struct in behavior.go
type FullCreateMessageRequest struct {
	JSONRPC string                     `json:"jsonrpc"`
	ID      interface{}                `json:"id"`
	Method  string                     `json:"method"` // MUST be "sampling/createMessage"
	Params  CreateMessageRequestParams `json:"params"`
}

// CreateMessageResult is the response from sampling.
type CreateMessageResult struct {
	Meta       *Meta   `json:"_meta,omitempty"`
	Role       Role    `json:"role"`
	Content    Content `json:"content"`
	Model      string  `json:"model"`
	StopReason string  `json:"stopReason,omitempty"`
}
