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

	// Test JSONRPCRequest
	req := JSONRPCRequest{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		ID:             "req1",
		Method:         "test",
		Params:         map[string]interface{}{"p": "v"},
	}
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"req1"`) || !strings.Contains(string(b), `"test"`) {
		t.Errorf("expected request fields")
	}

	// Test JSONRPCResponse
	res := JSONRPCResponse{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		ID:             123,
		Result:         "success",
	}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"success"`) {
		t.Errorf("expected response fields")
	}

	errRes := JSONRPCResponse{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		ID:             123,
		Error:          &JSONRPCError{Code: -32600, Message: "error"},
	}
	b, _ = json.Marshal(errRes)
	if !strings.Contains(string(b), `-32600`) {
		t.Errorf("expected error code")
	}

	// Test Result
	baseResult := Result{Meta: &Meta{ProgressToken: "token456"}}
	b, _ = json.Marshal(baseResult)
	if !strings.Contains(string(b), `"token456"`) {
		t.Errorf("expected meta progress token in base result")
	}

	// Test JSONRPCNotification
	not := JSONRPCNotification{
		JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
		Method:         "notify",
	}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"notify"`) {
		t.Errorf("expected notification method")
	}
}
func TestClientServerMeta(t *testing.T) {
	// Types are simple wrappers around JSONRPC base classes, validating they exist.
	_ = ClientNotification{JSONRPCNotification{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
	_ = ClientRequest{JSONRPCRequest{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
	_ = ClientResult{JSONRPCResponse{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
	_ = ServerNotification{JSONRPCNotification{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
	_ = ServerRequest{JSONRPCRequest{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
	_ = ServerResult{JSONRPCResponse{JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"}}}
}

func TestProgressToken(t *testing.T) {
	var pt ProgressToken = "token"
	if pt != "token" {
		t.Errorf("ProgressToken test failed")
	}
}
