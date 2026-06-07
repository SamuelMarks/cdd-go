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
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"` // MUST be "notifications/message"
	Params  struct {
		Level  LoggingLevel `json:"level"`
		Logger string       `json:"logger,omitempty"`
		Data   interface{}  `json:"data"`
	} `json:"params"`
}

// SetLevelRequest requests that the server change its logging level.
type SetLevelRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"` // MUST be "logging/setLevel"
	Params  struct {
		Level LoggingLevel `json:"level"`
	} `json:"params"`
}
