// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"encoding/json"
	"fmt"
)

// CompletionReference represents a discriminated union of PromptReference ("ref/prompt") or ResourceReference ("ref/resource").
type CompletionReference struct {
	// PromptRef holds the prompt reference when Type is "ref/prompt".
	PromptRef *PromptReference
	// ResourceRef holds the resource reference when Type is "ref/resource".
	ResourceRef *ResourceReference
}

// MarshalJSON serializes CompletionReference as either a PromptReference or ResourceReference.
func (cr CompletionReference) MarshalJSON() ([]byte, error) {
	if cr.PromptRef != nil {
		return json.Marshal(cr.PromptRef)
	}
	if cr.ResourceRef != nil {
		return json.Marshal(cr.ResourceRef)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes CompletionReference based on the "type" discriminator field.
// Parameter data is the raw JSON bytes, and it returns any unmarshal error.
func (cr *CompletionReference) UnmarshalJSON(data []byte) error {
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return err
	}
	switch header.Type {
	case "ref/prompt":
		var pr PromptReference
		if err := json.Unmarshal(data, &pr); err != nil {
			return err
		}
		*cr = CompletionReference{PromptRef: &pr}
		return nil
	case "ref/resource":
		var rr ResourceReference
		if err := json.Unmarshal(data, &rr); err != nil {
			return err
		}
		*cr = CompletionReference{ResourceRef: &rr}
		return nil
	default:
		return fmt.Errorf("unsupported completion reference type: %q", header.Type)
	}
}

// CompleteRequest is a request to complete an argument.
type CompleteRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "completion/complete".
	Method string `json:"method"`
	// Params holds the reference and argument to complete.
	Params struct {
		// Ref identifies the prompt or resource being completed.
		Ref CompletionReference `json:"ref"`
		// Argument holds the argument name and partial value.
		Argument struct {
			// Name is the argument name.
			Name string `json:"name"`
			// Value is the current partial argument value.
			Value string `json:"value"`
		} `json:"argument"`
	} `json:"params"`
}

// CompleteResult is the response for a completion request.
type CompleteResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Completion holds the matched completion values and pagination info.
	Completion struct {
		// Values lists the suggested completion strings.
		Values []string `json:"values"`
		// Total is the optional total number of available matches.
		Total int `json:"total,omitempty"`
		// HasMore indicates whether additional completion values exist beyond Values.
		HasMore bool `json:"hasMore,omitempty"`
	} `json:"completion"`
}
