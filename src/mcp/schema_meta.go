// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"encoding/json"
	"fmt"
)

// Meta object included in requests and results.
type Meta struct {
	// ProgressToken is an optional token used to associate progress notifications with a request.
	ProgressToken ProgressToken `json:"progressToken,omitempty"`
}

// RequestID represents a JSON-RPC 2.0 request identifier (string or integer).
type RequestID any

// RequestId is an alias for RequestID for MCP schema compatibility.
type RequestId = RequestID

// ProgressToken represents an MCP progress token identifier (string or integer).
type ProgressToken any

// JSONRPCMessage is the base type for all JSON-RPC messages.
type JSONRPCMessage struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCRequest is a request message.
type JSONRPCRequest struct {
	JSONRPCMessage
	// ID is the unique request identifier.
	ID RequestId `json:"id"`
	// Method is the RPC method name.
	Method string `json:"method"`
	// Params holds the structured or raw request parameters.
	Params any `json:"params,omitempty"`
}

// JSONRPCResponse is a response message.
type JSONRPCResponse struct {
	JSONRPCMessage
	// ID is the request identifier matching the original request.
	ID RequestId `json:"id"`
	// Result holds the successful execution payload.
	Result any `json:"result,omitempty"`
	// Error holds the JSON-RPC error object if the call failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCNotification is a notification message.
type JSONRPCNotification struct {
	JSONRPCMessage
	// Method is the notification method name.
	Method string `json:"method"`
	// Params holds the structured or raw notification parameters.
	Params any `json:"params,omitempty"`
}

// JSONRPCError is an error object.
type JSONRPCError struct {
	// Code is the numeric JSON-RPC error code.
	Code int `json:"code"`
	// Message is a short description of the error.
	Message string `json:"message"`
	// Data contains additional information about the error.
	Data any `json:"data,omitempty"`
}

// Validate checks that the JSONRPCMessage has version "2.0".
// It returns an error if the protocol version is invalid.
func (m JSONRPCMessage) Validate() error {
	if m.JSONRPC != "2.0" {
		return fmt.Errorf("expected jsonrpc 2.0, got %q", m.JSONRPC)
	}
	return nil
}

// ClientNotification represents a notification sent by a client.
type ClientNotification struct {
	JSONRPCNotification
}

// ClientRequest represents a request sent by a client.
type ClientRequest struct {
	JSONRPCRequest
}

// ClientResult represents a result sent by a client.
type ClientResult struct {
	JSONRPCResponse
}

// ServerNotification represents a notification sent by a server.
type ServerNotification struct {
	JSONRPCNotification
}

// ServerRequest represents a request sent by a server.
type ServerRequest struct {
	JSONRPCRequest
}

// ServerResult represents a result sent by a server.
type ServerResult struct {
	JSONRPCResponse
}

// Result is the base type for all results.
type Result struct {
	// Meta contains optional protocol metadata.
	Meta *Meta `json:"_meta,omitempty"`
}

// UnmarshalRequestID unmarshals a JSON value into a validated RequestID (string or int64).
// Parameter data is the raw JSON bytes, and it returns the parsed RequestID and error.
func UnmarshalRequestID(data []byte) (RequestID, error) {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return s, nil
	}
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		return n, nil
	}
	return nil, fmt.Errorf("invalid RequestID: must be string or integer, got %s", string(data))
}
