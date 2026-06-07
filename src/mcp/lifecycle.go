package mcp

// InitializeRequest represents the request for the initialize handshake.
type InitializeRequest struct {
	JSONRPC string                  `json:"jsonrpc"`
	ID      interface{}             `json:"id"`
	Method  string                  `json:"method"` // MUST be "initialize"
	Params  InitializeRequestParams `json:"params"`
}

// InitializeRequestParams represents the parameters for the initialize handshake.
type InitializeRequestParams struct {
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// InitializeResult represents the result of the initialize handshake.
type InitializeResult struct {
	Meta            *Meta       `json:"_meta,omitempty"`
	ProtocolVersion string      `json:"protocolVersion"`
	Capabilities    interface{} `json:"capabilities"`
	ServerInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
	Instructions string `json:"instructions,omitempty"`
}

// InitializedNotification represents the initialized acknowledgment.
type InitializedNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// PingRequest represents a liveness check.
type PingRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// CancelledNotification represents a request cancellation.
type CancelledNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  struct {
		RequestID interface{} `json:"requestId"`
		Reason    string      `json:"reason,omitempty"`
	} `json:"params"`
}
