package mcp

// CompleteRequest is a request to complete an argument.
type CompleteRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "completion/complete"
	Params  struct {
		Ref      interface{} `json:"ref"`
		Argument struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"argument"`
	} `json:"params"`
}

// CompleteResult is the response for a completion request.
type CompleteResult struct {
	Meta       *Meta `json:"_meta,omitempty"`
	Completion struct {
		Values  []string `json:"values"`
		Total   int      `json:"total,omitempty"`
		HasMore bool     `json:"hasMore,omitempty"`
	} `json:"completion"`
}
