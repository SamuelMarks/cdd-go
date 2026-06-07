package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaLoggingStructs(t *testing.T) {
	// Test LoggingLevel constants
	if LogLevelDebug != "debug" || LogLevelInfo != "info" || LogLevelNotice != "notice" ||
		LogLevelWarning != "warning" || LogLevelError != "error" || LogLevelCritical != "critical" ||
		LogLevelAlert != "alert" || LogLevelEmergency != "emergency" {
		t.Errorf("expected correct logging levels")
	}

	// Test LoggingMessageNotification
	not := LoggingMessageNotification{Method: "notifications/message"}
	not.Params.Level = LogLevelInfo
	not.Params.Logger = "system"
	not.Params.Data = "booting"
	b, _ := json.Marshal(not)
	if !strings.Contains(string(b), `"notifications/message"`) || !strings.Contains(string(b), `"info"`) || !strings.Contains(string(b), `"system"`) || !strings.Contains(string(b), `"booting"`) {
		t.Errorf("expected logging notification fields")
	}

	// Test SetLevelRequest
	req := SetLevelRequest{Method: "logging/setLevel"}
	req.Params.Level = LogLevelDebug
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"logging/setLevel"`) || !strings.Contains(string(b), `"debug"`) {
		t.Errorf("expected set level request fields")
	}
}
