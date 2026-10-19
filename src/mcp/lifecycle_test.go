package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestLifecycleStructsAndRoundTrip(t *testing.T) {
	reqParams := InitializeRequestParams{
		ProtocolVersion: "1.0",
		Capabilities:    ClientCapabilities{Sampling: &struct{}{}},
		ClientInfo:      Implementation{Name: "TestClient", Version: "0.1"},
	}

	initReq := InitializeRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "initialize", Params: reqParams}
	b, err := json.Marshal(initReq)
	if err != nil || !strings.Contains(string(b), `"TestClient"`) {
		t.Fatalf("expected TestClient in json: %v", err)
	}
	var initReqOut InitializeRequest
	if err := json.Unmarshal(b, &initReqOut); err != nil || !reflect.DeepEqual(initReq, initReqOut) {
		t.Errorf("InitializeRequest round-trip mismatch: %+v vs %+v (%v)", initReq, initReqOut, err)
	}

	initRes := InitializeResult{
		ProtocolVersion: "1.0",
		Capabilities:    ServerCapabilities{Logging: &struct{}{}},
		ServerInfo:      Implementation{Name: "TestServer", Version: "0.2"},
		Instructions:    "Use wisely",
	}
	b, err = json.Marshal(initRes)
	if err != nil || !strings.Contains(string(b), `"TestServer"`) {
		t.Fatalf("expected TestServer in json: %v", err)
	}
	var initResOut InitializeResult
	if err := json.Unmarshal(b, &initResOut); err != nil || !reflect.DeepEqual(initRes, initResOut) {
		t.Errorf("InitializeResult round-trip mismatch: %+v vs %+v (%v)", initRes, initResOut, err)
	}

	initNot := InitializedNotification{JSONRPC: "2.0", Method: "notifications/initialized"}
	b, err = json.Marshal(initNot)
	if err != nil || !strings.Contains(string(b), `"notifications/initialized"`) {
		t.Fatalf("expected notifications/initialized: %v", err)
	}

	ping := PingRequest{JSONRPC: "2.0", ID: NewIntRequestID(2), Method: "ping"}
	b, err = json.Marshal(ping)
	if err != nil || !strings.Contains(string(b), `"ping"`) {
		t.Fatalf("expected ping: %v", err)
	}

	cancel := CancelledNotification{JSONRPC: "2.0", Method: "$/cancelRequest"}
	cancel.Params.RequestID = NewIntRequestID(1)
	cancel.Params.Reason = "user cancelled"
	b, err = json.Marshal(cancel)
	if err != nil || !strings.Contains(string(b), `"requestId":1`) {
		t.Fatalf("expected request id: %v, %s", err, string(b))
	}
	var cancelOut CancelledNotification
	if err := json.Unmarshal(b, &cancelOut); err != nil || !reflect.DeepEqual(cancel, cancelOut) {
		t.Errorf("CancelledNotification round-trip mismatch: %+v vs %+v (%v)", cancel, cancelOut, err)
	}
}
