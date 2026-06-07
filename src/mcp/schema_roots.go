package mcp

// Root describes a single root object.
type Root struct {
	URI  string `json:"uri"`
	Name string `json:"name,omitempty"`
}

// ListRootsRequest requests a list of roots.
type ListRootsRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "roots/list"
	Params  struct {
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}

// ListRootsResult returns a list of roots.
type ListRootsResult struct {
	Meta  *Meta  `json:"_meta,omitempty"`
	Roots []Root `json:"roots"`
}

// RootsListChangedNotification notifies the server that the roots list has changed.
type RootsListChangedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/roots/list_changed"
	Params  struct {
		Meta *Meta `json:"_meta,omitempty"`
	} `json:"params,omitempty"`
}
