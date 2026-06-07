package mcp

// Resource represents a single resource object.
type Resource struct {
	Annotated
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
	Size        int64  `json:"size,omitempty"`
}

// ResourceTemplate describes a parameterized resource.
type ResourceTemplate struct {
	Annotated
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// ResourceReference points to an external resource.
type ResourceReference struct {
	Type string `json:"type"` // MUST be "ref/resource"
	URI  string `json:"uri"`
}

// ReadResourceRequest requests the content of a resource.
type ReadResourceRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "resources/read"
	Params  struct {
		URI string `json:"uri"`
	} `json:"params"`
}

// ReadResourceResult provides the content of a requested resource.
type ReadResourceResult struct {
	Meta     *Meta              `json:"_meta,omitempty"`
	Contents []ResourceContents `json:"contents"`
}

// ListResourcesRequest requests a list of available resources.
type ListResourcesRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "resources/list"
	Params  struct {
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListResourcesResult provides a paginated list of resources.
type ListResourcesResult struct {
	Meta       *Meta      `json:"_meta,omitempty"`
	NextCursor Cursor     `json:"nextCursor,omitempty"`
	Resources  []Resource `json:"resources"`
}

// ListResourceTemplatesRequest requests a list of resource templates.
type ListResourceTemplatesRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "resources/templates/list"
	Params  struct {
		Cursor Cursor `json:"cursor,omitempty"`
	} `json:"params,omitempty"`
}

// ListResourceTemplatesResult provides a paginated list of resource templates.
type ListResourceTemplatesResult struct {
	Meta              *Meta              `json:"_meta,omitempty"`
	NextCursor        Cursor             `json:"nextCursor,omitempty"`
	ResourceTemplates []ResourceTemplate `json:"resourceTemplates"`
}

// ResourceListChangedNotification notifies clients of changes to the resource list.
type ResourceListChangedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/resources/list_changed"
	Params  struct {
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ResourceUpdatedNotification notifies subscribers that a specific resource was updated.
type ResourceUpdatedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/resources/updated"
	Params  struct {
		URI string `json:"uri"`
	} `json:"params"`
}

// SubscribeRequest subscribes to updates for a specific resource.
type SubscribeRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "resources/subscribe"
	Params  struct {
		URI string `json:"uri"`
	} `json:"params"`
}

// UnsubscribeRequest unsubscribes from a resource.
type UnsubscribeRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "resources/unsubscribe"
	Params  struct {
		URI string `json:"uri"`
	} `json:"params"`
}
