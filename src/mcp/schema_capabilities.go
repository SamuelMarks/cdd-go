package mcp

// ClientCapabilities represents the capabilities of a client.
type ClientCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Roots        *struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"roots,omitempty"`
	Sampling *struct{} `json:"sampling,omitempty"`
}

// ServerCapabilities represents the capabilities of a server.
type ServerCapabilities struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Logging      *struct{}              `json:"logging,omitempty"`
	Prompts      *struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"prompts,omitempty"`
	Resources *struct {
		ListChanged bool `json:"listChanged,omitempty"`
		Subscribe   bool `json:"subscribe,omitempty"`
	} `json:"resources,omitempty"`
	Tools *struct {
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"tools,omitempty"`
}
