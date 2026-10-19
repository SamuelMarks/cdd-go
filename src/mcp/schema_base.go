// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"encoding/json"
	"fmt"
)

// Annotations provides optional metadata.
type Annotations struct {
	// Audience lists the intended recipients ("user", "assistant").
	Audience []string `json:"audience,omitempty"`
	// Priority indicates the relative importance of the content (0.0 to 1.0).
	Priority float64 `json:"priority,omitempty"`
}

// Annotated is a base for objects that can have annotations.
type Annotated struct {
	// Annotations holds optional metadata annotations.
	Annotations *Annotations `json:"annotations,omitempty"`
}

// BlobResourceContents represents binary resource contents.
type BlobResourceContents struct {
	// URI is the unique resource identifier.
	URI string `json:"uri"`
	// MimeType is the optional MIME type of the binary blob.
	MimeType string `json:"mimeType,omitempty"`
	// Blob is the base64-encoded binary data.
	Blob string `json:"blob"`
}

// IsResourceContents returns the resource contents kind ("blob").
func (BlobResourceContents) IsResourceContents() string { return "blob" }

// TextResourceContents represents text resource contents.
type TextResourceContents struct {
	// URI is the unique resource identifier.
	URI string `json:"uri"`
	// MimeType is the optional MIME type of the text.
	MimeType string `json:"mimeType,omitempty"`
	// Text is the textual content of the resource.
	Text string `json:"text"`
}

// IsResourceContents returns the resource contents kind ("text").
func (TextResourceContents) IsResourceContents() string { return "text" }

// ResourceContents can be either BlobResourceContents or TextResourceContents.
type ResourceContents interface {
	// IsResourceContents returns the sealed resource contents kind ("blob" or "text").
	IsResourceContents() string
}

// UnmarshalResourceContents unmarshals raw JSON into a concrete TextResourceContents or BlobResourceContents.
// Parameter data is the raw JSON payload, and it returns the decoded ResourceContents and error.
func UnmarshalResourceContents(data []byte) (ResourceContents, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if _, ok := raw["blob"]; ok {
		var b BlobResourceContents
		if err := json.Unmarshal(data, &b); err != nil {
			return nil, err
		}
		return b, nil
	}
	if _, ok := raw["text"]; ok {
		var t TextResourceContents
		if err := json.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, fmt.Errorf("unknown resource contents payload: %s", string(data))
}

// EmbeddedResource represents a resource embedded directly within a message.
type EmbeddedResource struct {
	Annotated
	// Type must be "resource".
	Type string `json:"type"`
	// Resource holds the embedded text or blob resource contents.
	Resource ResourceContents `json:"resource"`
}

// IsContent returns the content block type ("resource").
func (EmbeddedResource) IsContent() string { return "resource" }

// UnmarshalJSON deserializes an EmbeddedResource and its polymorphic ResourceContents field.
// Parameter data is the raw JSON bytes, and it returns any unmarshal error.
func (er *EmbeddedResource) UnmarshalJSON(data []byte) error {
	var raw struct {
		Annotated
		Type     string          `json:"type"`
		Resource json.RawMessage `json:"resource"`
	}
	_ = json.Unmarshal(data, &raw)
	rc, err := UnmarshalResourceContents(raw.Resource)
	if err != nil {
		return err
	}
	er.Annotated = raw.Annotated
	er.Type = raw.Type
	er.Resource = rc
	return nil
}

// TextContent represents text content within a message.
type TextContent struct {
	Annotated
	// Type must be "text".
	Type string `json:"type"`
	// Text is the message text.
	Text string `json:"text"`
}

// IsContent returns the content block type ("text").
func (TextContent) IsContent() string { return "text" }

// ImageContent represents image content within a message.
type ImageContent struct {
	Annotated
	// Type must be "image".
	Type string `json:"type"`
	// Data is the base64-encoded image payload.
	Data string `json:"data"`
	// MimeType is the image MIME type (e.g., "image/png").
	MimeType string `json:"mimeType"`
}

// IsContent returns the content block type ("image").
func (ImageContent) IsContent() string { return "image" }

// Content can be TextContent, ImageContent, or EmbeddedResource.
type Content interface {
	// IsContent returns the sealed content block type ("text", "image", or "resource").
	IsContent() string
}

// UnmarshalContent unmarshals raw JSON into a concrete TextContent, ImageContent, or EmbeddedResource.
// Parameter data is the raw JSON payload, and it returns the decoded Content and error.
func UnmarshalContent(data []byte) (Content, error) {
	var header struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return nil, err
	}
	switch header.Type {
	case "text":
		var tc TextContent
		if err := json.Unmarshal(data, &tc); err != nil {
			return nil, err
		}
		return tc, nil
	case "image":
		var ic ImageContent
		if err := json.Unmarshal(data, &ic); err != nil {
			return nil, err
		}
		return ic, nil
	case "resource":
		var er EmbeddedResource
		if err := json.Unmarshal(data, &er); err != nil {
			return nil, err
		}
		return er, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %q", header.Type)
	}
}

// Implementation describes the name and version of an MCP implementation.
type Implementation struct {
	// Name is the implementation name.
	Name string `json:"name"`
	// Version is the implementation version string.
	Version string `json:"version"`
}

// Role represents a conversational role (e.g., "user", "assistant").
type Role string

// RoleUser represents a user role.
const RoleUser Role = "user"

// RoleAssistant represents an assistant role.
const RoleAssistant Role = "assistant"

// EmptyResult represents an empty result.
type EmptyResult struct{}
