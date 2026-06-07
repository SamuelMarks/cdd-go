package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	data := []byte(`{"jsonrpc": "2.0", "method": "test", "id": 1}`)
	msg, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.JSONRPC != "2.0" {
		t.Errorf("expected 2.0")
	}
	if msg.Method != "test" {
		t.Errorf("expected test")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	data := []byte(`{invalid`)
	_, err := Parse(data)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseInvalidVersion(t *testing.T) {
	data := []byte(`{"jsonrpc": "1.0", "method": "test", "id": 1}`)
	_, err := Parse(data)
	if err == nil || !strings.Contains(err.Error(), "invalid jsonrpc version") {
		t.Errorf("expected version error, got %v", err)
	}
}

func TestErrorString(t *testing.T) {
	err := &Error{Code: ErrMethodNotFound, Message: "Method not found"}
	if err.Error() != "jsonrpc error -32601: Method not found" {
		t.Errorf("unexpected error string: %v", err.Error())
	}
}

func TestMapErrorCode(t *testing.T) {
	cases := []struct {
		code     int
		details  string
		expected string
	}{
		{ErrParse, "", "Parse error"},
		{ErrInvalidRequest, "missing id", "Invalid Request: missing id"},
		{ErrMethodNotFound, "", "Method not found"},
		{ErrInvalidParams, "", "Invalid params"},
		{ErrInternal, "", "Internal error"},
		{-12345, "", "Unknown error"},
	}

	for _, c := range cases {
		err := MapErrorCode(c.code, c.details)
		if err.Code != c.code {
			t.Errorf("expected code %d, got %d", c.code, err.Code)
		}
		if err.Message != c.expected {
			t.Errorf("expected message %s, got %s", c.expected, err.Message)
		}
	}
}

func TestStructFields(t *testing.T) {
	req := Request{JSONRPC: "2.0", ID: 1, Method: "test"}
	b, _ := json.Marshal(req)
	if !strings.Contains(string(b), `"id":1`) {
		t.Errorf("expected id in request")
	}

	res := Response{JSONRPC: "2.0", ID: 1, Result: "ok"}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"result":"ok"`) {
		t.Errorf("expected result in response")
	}

	not := Notification{JSONRPC: "2.0", Method: "notify"}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"method":"notify"`) {
		t.Errorf("expected method in notification")
	}
}
