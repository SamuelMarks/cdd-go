// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Meta object included in requests and results.
type Meta struct {
	// ProgressToken is an optional token used to associate progress notifications with a request.
	ProgressToken *ProgressToken `json:"progressToken,omitempty"`
}

// RequestID represents a strongly-typed JSON-RPC 2.0 request identifier (string or integer).
type RequestID struct {
	// StrVal holds a string request identifier when present.
	StrVal *string
	// IntVal holds an integer request identifier when present.
	IntVal *int64
}

// RequestId is an alias for RequestID for MCP schema compatibility.
type RequestId = RequestID

// NewStringRequestID constructs a RequestID from a string value.
// Parameter s is the string request identifier, and it returns the initialized RequestID.
func NewStringRequestID(s string) RequestID {
	return RequestID{StrVal: &s}
}

// NewIntRequestID constructs a RequestID from an int64 value.
// Parameter n is the integer request identifier, and it returns the initialized RequestID.
func NewIntRequestID(n int64) RequestID {
	return RequestID{IntVal: &n}
}

// IsZero reports whether the RequestID has neither a string nor an integer value set.
func (id RequestID) IsZero() bool {
	return id.StrVal == nil && id.IntVal == nil
}

// String returns the canonical string representation of the RequestID.
func (id RequestID) String() string {
	if id.StrVal != nil {
		return *id.StrVal
	}
	if id.IntVal != nil {
		return strconv.FormatInt(*id.IntVal, 10)
	}
	return ""
}

// MarshalJSON serializes RequestID as a JSON string, integer, or null.
func (id RequestID) MarshalJSON() ([]byte, error) {
	if id.StrVal != nil {
		return json.Marshal(*id.StrVal)
	}
	if id.IntVal != nil {
		return json.Marshal(*id.IntVal)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes a JSON string, integer, or null into RequestID.
// Parameter data is the raw JSON bytes, and it returns any unmarshal error.
func (id *RequestID) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*id = RequestID{}
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err == nil {
			*id = NewStringRequestID(s)
			return nil
		}
	}
	if (trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-' {
		var n int64
		if err := json.Unmarshal(trimmed, &n); err == nil {
			*id = NewIntRequestID(n)
			return nil
		}
	}
	return fmt.Errorf("invalid RequestID: must be string or integer, got %s", string(data))
}

// UnmarshalRequestID unmarshals a JSON value into a validated RequestID (string or int64).
// Parameter data is the raw JSON bytes, and it returns the parsed RequestID and error.
func UnmarshalRequestID(data []byte) (RequestID, error) {
	var id RequestID
	if err := json.Unmarshal(data, &id); err != nil {
		return RequestID{}, err
	}
	if id.IsZero() {
		return RequestID{}, fmt.Errorf("invalid RequestID: must be string or integer, got %s", string(data))
	}
	return id, nil
}

// ProgressToken represents a strongly-typed MCP progress token identifier (string or integer).
type ProgressToken struct {
	// StrVal holds a string progress token when present.
	StrVal *string
	// IntVal holds an integer progress token when present.
	IntVal *int64
}

// NewStringProgressToken constructs a ProgressToken from a string value.
// Parameter s is the string token, and it returns the initialized ProgressToken.
func NewStringProgressToken(s string) ProgressToken {
	return ProgressToken{StrVal: &s}
}

// NewIntProgressToken constructs a ProgressToken from an int64 value.
// Parameter n is the integer token, and it returns the initialized ProgressToken.
func NewIntProgressToken(n int64) ProgressToken {
	return ProgressToken{IntVal: &n}
}

// String returns the canonical string representation of the ProgressToken.
func (pt ProgressToken) String() string {
	if pt.StrVal != nil {
		return *pt.StrVal
	}
	if pt.IntVal != nil {
		return strconv.FormatInt(*pt.IntVal, 10)
	}
	return ""
}

// MarshalJSON serializes ProgressToken as a JSON string, integer, or null.
func (pt ProgressToken) MarshalJSON() ([]byte, error) {
	if pt.StrVal != nil {
		return json.Marshal(*pt.StrVal)
	}
	if pt.IntVal != nil {
		return json.Marshal(*pt.IntVal)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes a JSON string or integer into ProgressToken.
// Parameter data is the raw JSON bytes, and it returns any unmarshal error.
func (pt *ProgressToken) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*pt = ProgressToken{}
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err == nil {
			*pt = NewStringProgressToken(s)
			return nil
		}
	}
	if (trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-' {
		var n int64
		if err := json.Unmarshal(trimmed, &n); err == nil {
			*pt = NewIntProgressToken(n)
			return nil
		}
	}
	return fmt.Errorf("invalid ProgressToken: must be string or integer, got %s", string(data))
}

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
	// Params holds the raw JSON request parameters.
	Params json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse is a response message.
type JSONRPCResponse struct {
	JSONRPCMessage
	// ID is the request identifier matching the original request.
	ID RequestId `json:"id"`
	// Result holds the raw JSON execution payload.
	Result json.RawMessage `json:"result,omitempty"`
	// Error holds the JSON-RPC error object if the call failed.
	Error *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCNotification is a notification message.
type JSONRPCNotification struct {
	JSONRPCMessage
	// Method is the notification method name.
	Method string `json:"method"`
	// Params holds the raw JSON notification parameters.
	Params json.RawMessage `json:"params,omitempty"`
}

// JSONRPCError is an error object.
type JSONRPCError struct {
	// Code is the numeric JSON-RPC error code.
	Code int `json:"code"`
	// Message is a short description of the error.
	Message string `json:"message"`
	// Data contains additional raw JSON information about the error.
	Data json.RawMessage `json:"data,omitempty"`
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
