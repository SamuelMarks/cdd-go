package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaMetaStructs(t *testing.T) {
	// Test Meta
	meta := Meta{ProgressToken: "token123"}
	b, _ := json.Marshal(meta)
	if !strings.Contains(string(b), `"token123"`) {
		t.Errorf("expected progress token")
	}

	// Test JSONRPCRequest, Response, Notification, Error
	req := JSONRPCRequest{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		ID:             1,
		Method:         "test",
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid JSONRPCMessage, got %v", err)
	}
	invalidMsg := JSONRPCMessage{JSONRPC: "1.0"}
	if err := invalidMsg.Validate(); err == nil {
		t.Errorf("expected error for invalid JSONRPC version")
	}
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"method":"test"`) {
		t.Errorf("expected request method")
	}

	resp := JSONRPCResponse{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		ID:             "1",
		Error:          &JSONRPCError{Code: -32600, Message: "Invalid"},
	}
	b, _ = json.Marshal(resp)
	if !strings.Contains(string(b), `"-32600"`) && !strings.Contains(string(b), `-32600`) {
		t.Errorf("expected error code")
	}

	not := JSONRPCNotification{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		Method:         "notify",
	}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"notify"`) {
		t.Errorf("expected notify method")
	}

	res := Result{Meta: &meta}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"token123"`) {
		t.Errorf("expected meta in result")
	}

	_ = ClientNotification{not}
	_ = ClientRequest{req}
	_ = ClientResult{resp}
	_ = ServerNotification{not}
	_ = ServerRequest{req}
	_ = ServerResult{resp}
}

func TestProgressTokenAndRequestID(t *testing.T) {
	var pt ProgressToken = "token"
	if pt != "token" {
		t.Errorf("ProgressToken test failed")
	}

	idStr, err := UnmarshalRequestID([]byte(`"req-1"`))
	if err != nil || idStr != "req-1" {
		t.Errorf("expected req-1, got %v, %v", idStr, err)
	}
	idInt, err := UnmarshalRequestID([]byte(`42`))
	if err != nil || idInt != int64(42) {
		t.Errorf("expected 42, got %v, %v", idInt, err)
	}
	if _, err := UnmarshalRequestID([]byte(`true`)); err == nil {
		t.Errorf("expected error for boolean RequestID")
	}
}
