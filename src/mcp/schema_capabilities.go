// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import "encoding/json"

// ClientCapabilities represents the capabilities of a client.
type ClientCapabilities struct {
	// Experimental holds non-standard experimental capability flags as raw JSON objects.
	Experimental map[string]json.RawMessage `json:"experimental,omitempty"`
	// Roots indicates whether the client supports workspace root listing.
	Roots *struct {
		// ListChanged indicates whether the client emits root list change notifications.
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"roots,omitempty"`
	// Sampling indicates whether the client supports LLM sampling requests.
	Sampling *struct{} `json:"sampling,omitempty"`
}

// ServerCapabilities represents the capabilities of a server.
type ServerCapabilities struct {
	// Experimental holds non-standard experimental capability flags as raw JSON objects.
	Experimental map[string]json.RawMessage `json:"experimental,omitempty"`
	// Logging indicates whether the server supports structured log notifications.
	Logging *struct{} `json:"logging,omitempty"`
	// Prompts indicates whether the server offers prompt templates.
	Prompts *struct {
		// ListChanged indicates whether the server emits prompt list change notifications.
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"prompts,omitempty"`
	// Resources indicates whether the server exposes readable resources.
	Resources *struct {
		// ListChanged indicates whether the server emits resource list change notifications.
		ListChanged bool `json:"listChanged,omitempty"`
		// Subscribe indicates whether the server supports resource subscriptions.
		Subscribe bool `json:"subscribe,omitempty"`
	} `json:"resources,omitempty"`
	// Tools indicates whether the server exposes callable tools.
	Tools *struct {
		// ListChanged indicates whether the server emits tool list change notifications.
		ListChanged bool `json:"listChanged,omitempty"`
	} `json:"tools,omitempty"`
}
