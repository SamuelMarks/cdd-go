package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLifecycleStructs(t *testing.T) {
	// Test InitializeRequestParams
	reqParams := InitializeRequestParams{ProtocolVersion: "1.0"}
	reqParams.ClientInfo.Name = "testClient"
	reqParams.ClientInfo.Version = "1.0.0"

	// Test InitializeRequest
	req := InitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  reqParams,
	}
	b, _ := json.Marshal(req)
	if !strings.Contains(string(b), `"initialize"`) {
		t.Errorf("expected initialize method")
	}

	// Test InitializeResult
	res := InitializeResult{ProtocolVersion: "1.0", Instructions: "do things"}
	res.ServerInfo.Name = "testServer"
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"testServer"`) {
		t.Errorf("expected server name")
	}
	if !strings.Contains(string(b), `"instructions":"do things"`) {
		t.Errorf("expected instructions")
	}

	// Test InitializedNotification
	not := InitializedNotification{JSONRPC: "2.0", Method: "initialized"}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"initialized"`) {
		t.Errorf("expected method")
	}

	// Test PingRequest
	ping := PingRequest{JSONRPC: "2.0", Method: "ping", ID: 1}
	b, _ = json.Marshal(ping)
	if !strings.Contains(string(b), `"ping"`) {
		t.Errorf("expected ping method")
	}

	// Test CancelledNotification
	cancel := CancelledNotification{JSONRPC: "2.0", Method: "$/cancelRequest"}
	cancel.Params.RequestID = 1
	cancel.Params.Reason = "user cancelled"
	b, _ = json.Marshal(cancel)
	if !strings.Contains(string(b), `"requestId":1`) {
		t.Errorf("expected request id")
	}
}
