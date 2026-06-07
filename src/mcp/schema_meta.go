package mcp

// Meta object included in requests and results.
type Meta struct {
	ProgressToken interface{} `json:"progressToken,omitempty"`
}

// RequestId is either a string or an integer.
type RequestId interface{}

// JSONRPCMessage is the base type for all JSON-RPC messages.
type JSONRPCMessage struct {
	JSONRPC string `json:"jsonrpc"`
}

// JSONRPCRequest is a request message.
type JSONRPCRequest struct {
	JSONRPCMessage
	ID     RequestId   `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// JSONRPCResponse is a response message.
type JSONRPCResponse struct {
	JSONRPCMessage
	ID     RequestId     `json:"id"`
	Result interface{}   `json:"result,omitempty"`
	Error  *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCNotification is a notification message.
type JSONRPCNotification struct {
	JSONRPCMessage
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// JSONRPCError is an error object.
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
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
	Meta *Meta `json:"_meta,omitempty"`
}

// ProgressToken is either a string or an integer.
type ProgressToken interface{}
