package mcp

// Annotations provides optional metadata.
type Annotations struct {
	Audience []string `json:"audience,omitempty"`
	Priority float64  `json:"priority,omitempty"`
}

// Annotated is a base for objects that can have annotations.
type Annotated struct {
	Annotations *Annotations `json:"annotations,omitempty"`
}

// BlobResourceContents represents binary resource contents.
type BlobResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Blob     string `json:"blob"` // Base64 encoded
}

// TextResourceContents represents text resource contents.
type TextResourceContents struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text"`
}

// ResourceContents can be either BlobResourceContents or TextResourceContents
type ResourceContents interface{}

// EmbeddedResource represents a resource embedded directly within a message.
type EmbeddedResource struct {
	Annotated
	Type     string           `json:"type"` // MUST be "resource"
	Resource ResourceContents `json:"resource"`
}

// TextContent represents text content within a message.
type TextContent struct {
	Annotated
	Type string `json:"type"` // MUST be "text"
	Text string `json:"text"`
}

// ImageContent represents image content within a message.
type ImageContent struct {
	Annotated
	Type     string `json:"type"` // MUST be "image"
	Data     string `json:"data"` // Base64 encoded
	MimeType string `json:"mimeType"`
}

// Content can be TextContent, ImageContent, or EmbeddedResource
type Content interface{}

// Implementation describes the name and version of an MCP implementation.
type Implementation struct {
	Name    string `json:"name"`
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
