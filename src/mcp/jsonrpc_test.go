package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseAndMessageRoundTrip(t *testing.T) {
	data := []byte(`{"jsonrpc":"2.0","method":"test","id":1}`)
	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Method != "test" || msg.ID.IntVal == nil || *msg.ID.IntVal != 1 {
		t.Errorf("expected method test and id 1, got %+v", msg)
	}

	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	msgRT, err := Parse(b)
	if err != nil || msgRT.Method != "test" || msgRT.ID.String() != "1" {
		t.Errorf("round-trip mismatch: %+v, %v", msgRT, err)
	}

	// Test invalid JSON
	_, err = Parse([]byte(`{invalid`))
	if err == nil {
		t.Errorf("expected error for invalid json")
	}

	// Test invalid version
	_, err = Parse([]byte(`{"jsonrpc":"1.0","method":"test"}`))
	if err == nil {
		t.Errorf("expected error for invalid version")
	}
}

func TestErrorAndMapErrorCode(t *testing.T) {
	err := MapErrorCode(ErrParse, "unexpected EOF")
	if err.Code != ErrParse || err.Message != "Parse error: unexpected EOF" {
		t.Errorf("unexpected error mapping: %v", err)
	}
	if err.Error() != "jsonrpc error -32700: Parse error: unexpected EOF" {
		t.Errorf("unexpected Error() string: %v", err.Error())
	}

	err = MapErrorCode(ErrInvalidRequest, "")
	if err.Message != "Invalid Request" {
		t.Errorf("unexpected message: %v", err.Message)
	}
	err = MapErrorCode(ErrMethodNotFound, "")
	if err.Message != "Method not found" {
		t.Errorf("unexpected message: %v", err.Message)
	}
	err = MapErrorCode(ErrInvalidParams, "")
	if err.Message != "Invalid params" {
		t.Errorf("unexpected message: %v", err.Message)
	}
	err = MapErrorCode(ErrInternal, "")
	if err.Message != "Internal error" {
		t.Errorf("unexpected message: %v", err.Message)
	}
	err = MapErrorCode(-32000, "")
	if err.Message != "Unknown error" {
		t.Errorf("unexpected message: %v", err.Message)
	}
}

func TestRequestResponseNotificationRoundTrip(t *testing.T) {
	req := Request{
		JSONRPC: "2.0",
		ID:      NewStringRequestID("req-1"),
		Method:  "ping",
		Params:  json.RawMessage(`{"k":"v"}`),
	}
	b, err := json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"method":"ping"`) {
		t.Fatalf("expected method in request: %v, %s", err, string(b))
	}
	var reqOut Request
	if err := json.Unmarshal(b, &reqOut); err != nil || reqOut.ID.String() != "req-1" || reqOut.Method != "ping" {
		t.Errorf("request round-trip failed: %+v, %v", reqOut, err)
	}

	resp := Response{
		JSONRPC: "2.0",
		ID:      NewIntRequestID(7),
		Result:  json.RawMessage(`"pong"`),
	}
	b, err = json.Marshal(resp)
	if err != nil || !strings.Contains(string(b), `"result":"pong"`) {
		t.Fatalf("expected result in response: %v", err)
	}
	var respOut Response
	if err := json.Unmarshal(b, &respOut); err != nil || respOut.ID.String() != "7" || string(respOut.Result) != `"pong"` {
		t.Errorf("response round-trip failed: %+v, %v", respOut, err)
	}

	not := Notification{JSONRPC: "2.0", Method: "notify", Params: json.RawMessage(`{}`)}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"method":"notify"`) {
		t.Fatalf("expected method in notification: %v", err)
	}
	var notOut Notification
	if err := json.Unmarshal(b, &notOut); err != nil || notOut.Method != "notify" {
		t.Errorf("notification round-trip failed: %+v, %v", notOut, err)
	}
}
