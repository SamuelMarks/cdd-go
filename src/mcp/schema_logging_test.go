package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaLoggingStructsAndRoundTrip(t *testing.T) {
	if LogLevelDebug != "debug" || LogLevelInfo != "info" || LogLevelNotice != "notice" ||
		LogLevelWarning != "warning" || LogLevelError != "error" || LogLevelCritical != "critical" ||
		LogLevelAlert != "alert" || LogLevelEmergency != "emergency" {
		t.Errorf("expected valid logging levels")
	}

	not := LoggingMessageNotification{JSONRPC: "2.0", Method: "notifications/message"}
	not.Params.Level = LogLevelInfo
	not.Params.Logger = "testLogger"
	not.Params.Data = json.RawMessage(`"test message"`)
	b, err := json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"testLogger"`) || !strings.Contains(string(b), `"info"`) {
		t.Fatalf("expected logging notification fields: %v", err)
	}
	var notOut LoggingMessageNotification
	if err := json.Unmarshal(b, &notOut); err != nil || notOut.Params.Logger != "testLogger" || string(notOut.Params.Data) != `"test message"` {
		t.Errorf("LoggingMessageNotification round-trip failed: %+v, %v", notOut, err)
	}

	req := SetLevelRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "logging/setLevel"}
	req.Params.Level = LogLevelDebug
	b, err = json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"logging/setLevel"`) || !strings.Contains(string(b), `"debug"`) {
		t.Fatalf("expected set level request fields: %v", err)
	}
	var reqOut SetLevelRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || reqOut.Params.Level != LogLevelDebug {
		t.Errorf("SetLevelRequest round-trip failed: %+v, %v", reqOut, err)
	}
}
