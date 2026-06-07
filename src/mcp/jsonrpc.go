package mcp

import (
	"encoding/json"
	"fmt"
)

// Standard JSON-RPC 2.0 Error Codes
const (
	ErrParse          = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
)

// Message is a raw JSON-RPC message.
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents a JSON-RPC error object.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// Request is a typed JSON-RPC request.
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response is a typed JSON-RPC response.
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Notification is a typed JSON-RPC notification.
type Notification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Parse parses a raw JSON payload into a Message envelope.
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
