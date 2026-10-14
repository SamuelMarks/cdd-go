// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// InitializeRequest represents the request for the initialize handshake.
type InitializeRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "initialize".
	Method string `json:"method"`
	// Params holds the initialization handshake parameters.
	Params InitializeRequestParams `json:"params"`
}

// InitializeRequestParams represents the parameters for the initialize handshake.
type InitializeRequestParams struct {
	// ProtocolVersion is the MCP protocol version requested by the client.
	ProtocolVersion string `json:"protocolVersion"`
	// Capabilities describes the client capabilities.
	Capabilities any `json:"capabilities"`
	// ClientInfo identifies the client application name and version.
	ClientInfo struct {
		// Name is the client name.
		Name string `json:"name"`
		// Version is the client version.
		Version string `json:"version"`
	} `json:"clientInfo"`
}

// InitializeResult represents the result of the initialize handshake.
type InitializeResult struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
	// ProtocolVersion is the MCP protocol version negotiated by the server.
	ProtocolVersion string `json:"protocolVersion"`
	// Capabilities describes the server capabilities.
	Capabilities any `json:"capabilities"`
	// ServerInfo identifies the server application name and version.
	ServerInfo struct {
		// Name is the server name.
		Name string `json:"name"`
		// Version is the server version.
		Version string `json:"version"`
	} `json:"serverInfo"`
	// Instructions provides optional usage instructions for the client.
	Instructions string `json:"instructions,omitempty"`
}

// InitializedNotification represents the initialized acknowledgment.
type InitializedNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is "notifications/initialized".
	Method string `json:"method"`
	// Params holds optional notification parameters.
	Params any `json:"params,omitempty"`
}

// PingRequest represents a liveness check.
type PingRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method is "ping".
	Method string `json:"method"`
	// Params holds optional ping parameters.
	Params any `json:"params,omitempty"`
}

// CancelledNotification represents a request cancellation.
type CancelledNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is "notifications/cancelled" or "$/cancelRequest".
	Method string `json:"method"`
	// Params holds the cancelled request ID and optional reason.
	Params struct {
		// RequestID is the identifier of the request to cancel.
		RequestID RequestID `json:"requestId"`
		// Reason is the optional human-readable cancellation reason.
		Reason string `json:"reason,omitempty"`
	} `json:"params"`
}
