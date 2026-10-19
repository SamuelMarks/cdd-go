// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"encoding/json"
	"fmt"
)

// Standard JSON-RPC 2.0 Error Codes.
const (
	// ErrParse indicates invalid JSON was received by the server.
	ErrParse = -32700
	// ErrInvalidRequest indicates the JSON sent is not a valid Request object.
	ErrInvalidRequest = -32600
	// ErrMethodNotFound indicates the method does not exist or is not available.
	ErrMethodNotFound = -32601
	// ErrInvalidParams indicates invalid method parameter(s).
	ErrInvalidParams = -32602
	// ErrInternal indicates an internal JSON-RPC error.
	ErrInternal = -32603
)

// Message is a raw JSON-RPC message.
type Message struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the optional request/response identifier.
	ID RequestID `json:"id,omitempty"`
	// Method is the optional RPC method name.
	Method string `json:"method,omitempty"`
	// Params holds the raw JSON parameters.
	Params json.RawMessage `json:"params,omitempty"`
	// Result holds the raw JSON result.
	Result json.RawMessage `json:"result,omitempty"`
	// Error holds the JSON-RPC error object if present.
	Error *Error `json:"error,omitempty"`
}

// Error represents a JSON-RPC error object.
type Error struct {
	// Code is the numeric JSON-RPC error code.
	Code int `json:"code"`
	// Message is a short description of the error.
	Message string `json:"message"`
	// Data contains optional additional raw JSON error details.
	Data json.RawMessage `json:"data,omitempty"`
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// Request is a typed JSON-RPC request.
type Request struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the request identifier.
	ID RequestID `json:"id"`
	// Method is the RPC method name.
	Method string `json:"method"`
	// Params holds the raw JSON request parameters.
	Params json.RawMessage `json:"params,omitempty"`
}

// Response is a typed JSON-RPC response.
type Response struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the response identifier matching the request.
	ID RequestID `json:"id"`
	// Result holds the raw JSON result payload.
	Result json.RawMessage `json:"result,omitempty"`
	// Error holds the error payload if the request failed.
	Error *Error `json:"error,omitempty"`
}

// Notification is a typed JSON-RPC notification.
type Notification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method is the notification method name.
	Method string `json:"method"`
	// Params holds the raw JSON notification parameters.
	Params json.RawMessage `json:"params,omitempty"`
}

// Parse parses a raw JSON payload into a Message envelope.
// Parameter data is the raw JSON bytes, and it returns the parsed Message and error.
func Parse(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	if msg.JSONRPC != "2.0" {
		return nil, fmt.Errorf("invalid jsonrpc version, expected 2.0")
	}
	return &msg, nil
}

// MapErrorCode provides a standard mapping to standard Error objects.
// Parameter code is the JSON-RPC error code, details is optional detail text, and it returns the constructed Error.
func MapErrorCode(code int, details string) *Error {
	msg := "Unknown error"
	switch code {
	case ErrParse:
		msg = "Parse error"
	case ErrInvalidRequest:
		msg = "Invalid Request"
	case ErrMethodNotFound:
		msg = "Method not found"
	case ErrInvalidParams:
		msg = "Invalid params"
	case ErrInternal:
		msg = "Internal error"
	}
	if details != "" {
		msg = msg + ": " + details
	}
	return &Error{Code: code, Message: msg}
}
