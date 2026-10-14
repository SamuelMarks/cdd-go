// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// Resource represents a single resource object.
type Resource struct {
	Annotated
	// URI is the unique resource location.
	URI string `json:"uri"`
	// Name is the human-readable resource name.
	Name string `json:"name"`
	// Description explains the purpose and contents of the resource.
	Description string `json:"description,omitempty"`
	// MimeType specifies the content MIME type.
	MimeType string `json:"mimeType,omitempty"`
	// Size is the optional byte size of the resource.
	Size int64 `json:"size,omitempty"`
}

// ResourceTemplate describes a parameterized resource.
type ResourceTemplate struct {
	Annotated
	// URITemplate is the RFC 6570 URI template string.
	URITemplate string `json:"uriTemplate"`
	// Name is the human-readable template name.
	Name string `json:"name"`
	// Description explains how the template is parameterized.
	Description string `json:"description,omitempty"`
	// MimeType specifies the MIME type of resources matching this template.
	MimeType string `json:"mimeType,omitempty"`
}

// ResourceReference points to an external resource.
type ResourceReference struct {
	// Type must be "ref/resource".
	Type string `json:"type"`
	// URI is the target resource URI.
	URI string `json:"uri"`
}

// ReadResourceRequest requests the content of a resource.
type ReadResourceRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "resources/read".
	Method string `json:"method"`
	// Params holds the resource URI to read.
	Params struct {
		// URI is the target resource URI.
		URI string `json:"uri"`
	} `json:"params"`
}

// ReadResourceResult provides the content of a requested resource.
type ReadResourceResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// Contents lists the text or blob resource content items.
	Contents []ResourceContents `json:"contents"`
}

// ListResourcesRequest requests a list of available resources.
type ListResourcesRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "resources/list".
	Method string `json:"method"`
	// Params holds optional pagination parameters.
	Params struct {
		// Cursor is the pagination cursor token.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListResourcesResult provides a paginated list of resources.
type ListResourcesResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// NextCursor is the opaque cursor for the next page of resources.
	NextCursor Cursor `json:"nextCursor,omitempty"`
	// Resources lists the available resources.
	Resources []Resource `json:"resources"`
}

// ListResourceTemplatesRequest requests a list of resource templates.
type ListResourceTemplatesRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "resources/templates/list".
	Method string `json:"method"`
	// Params holds optional pagination parameters.
	Params struct {
		// Cursor is the pagination cursor token.
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListResourceTemplatesResult provides a paginated list of resource templates.
type ListResourceTemplatesResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// NextCursor is the opaque cursor for the next page of templates.
	NextCursor Cursor `json:"nextCursor,omitempty"`
	// ResourceTemplates lists the available resource templates.
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates"`
}

// ResourceListChangedNotification notifies clients of changes to the resource list.
type ResourceListChangedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/resources/list_changed".
	Method string `json:"method"`
	// Params holds optional metadata.
	Params struct {
		// Meta contains optional protocol metadata.
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ResourceUpdatedNotification notifies subscribers that a specific resource was updated.
type ResourceUpdatedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/resources/updated".
	Method string `json:"method"`
	// Params identifies the updated resource URI.
	Params struct {
		// URI is the URI of the updated resource.
		URI string `json:"uri"`
	} `json:"params"`
}

// SubscribeRequest subscribes to updates for a specific resource.
type SubscribeRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "resources/subscribe".
	Method string `json:"method"`
	// Params identifies the resource URI to subscribe to.
	Params struct {
		// URI is the resource URI to watch.
		URI string `json:"uri"`
	} `json:"params"`
}

// UnsubscribeRequest unsubscribes from a resource.
type UnsubscribeRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "resources/unsubscribe".
	Method string `json:"method"`
	// Params identifies the resource URI to unsubscribe from.
	Params struct {
		// URI is the resource URI to stop watching.
		URI string `json:"uri"`
	} `json:"params"`
}
