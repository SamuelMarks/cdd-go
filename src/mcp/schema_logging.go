// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// LoggingLevel represents the severity of a log message.
type LoggingLevel string

// LogLevelDebug is for debugging.
const LogLevelDebug LoggingLevel = "debug"

// LogLevelInfo is for informational messages.
const LogLevelInfo LoggingLevel = "info"

// LogLevelNotice is for normal but significant conditions.
const LogLevelNotice LoggingLevel = "notice"

// LogLevelWarning is for warning conditions.
const LogLevelWarning LoggingLevel = "warning"

// LogLevelError is for error conditions.
const LogLevelError LoggingLevel = "error"

// LogLevelCritical is for critical conditions.
const LogLevelCritical LoggingLevel = "critical"

// LogLevelAlert is for action that must be taken immediately.
const LogLevelAlert LoggingLevel = "alert"

// LogLevelEmergency is for system is unusable conditions.
const LogLevelEmergency LoggingLevel = "emergency"

// LoggingMessageNotification sends a log message to the client.
type LoggingMessageNotification struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// Method must be "notifications/message".
	Method string `json:"method"`
	// Params holds the log level, logger name, and payload.
	Params struct {
		// Level is the severity level of the log entry.
		Level LoggingLevel `json:"level"`
		// Logger is the optional name of the logger emitting the message.
		Logger string `json:"logger,omitempty"`
		// Data is the arbitrary serializable log payload.
		Data any `json:"data"`
	} `json:"params"`
}

// SetLevelRequest requests that the server change its logging level.
type SetLevelRequest struct {
	// JSONRPC specifies the JSON-RPC protocol version ("2.0").
	JSONRPC string `json:"jsonrpc"`
	// ID is the unique request identifier.
	ID RequestID `json:"id"`
	// Method must be "logging/setLevel".
	Method string `json:"method"`
	// Params holds the requested minimum logging level.
	Params struct {
		// Level is the minimum LoggingLevel to emit.
		Level LoggingLevel `json:"level"`
	} `json:"params"`
}
