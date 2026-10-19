package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaMetaStructsAndRoundTrip(t *testing.T) {
	tok := NewStringProgressToken("token123")
	if tok.String() != "token123" {
		t.Errorf("expected token123, got %q", tok.String())
	}
	intTok := NewIntProgressToken(99)
	if intTok.String() != "99" {
		t.Errorf("expected 99, got %q", intTok.String())
	}
	var zeroTok ProgressToken
	if zeroTok.String() != "" {
		t.Errorf("expected empty string for zero token")
	}
	zeroBytes, err := json.Marshal(zeroTok)
	if err != nil || string(zeroBytes) != "null" {
		t.Errorf("expected null for zero ProgressToken, got %s, %v", string(zeroBytes), err)
	}
	if err := json.Unmarshal([]byte(`null`), &zeroTok); err != nil {
		t.Errorf("unexpected error unmarshaling null ProgressToken: %v", err)
	}
	if err := json.Unmarshal([]byte(`99`), &zeroTok); err != nil || zeroTok.String() != "99" {
		t.Errorf("expected 99 ProgressToken, got %v, %v", zeroTok, err)
	}
	if err := json.Unmarshal([]byte(`true`), &zeroTok); err == nil {
		t.Errorf("expected error for boolean ProgressToken")
	}

	meta := Meta{ProgressToken: &tok}
	b, err := json.Marshal(meta)
	if err != nil || !strings.Contains(string(b), `"progressToken":"token123"`) {
		t.Fatalf("expected progressToken: %v, %s", err, string(b))
	}
	var metaOut Meta
	if err := json.Unmarshal(b, &metaOut); err != nil || metaOut.ProgressToken == nil || metaOut.ProgressToken.String() != "token123" {
		t.Errorf("Meta round-trip failed: %+v, %v", metaOut, err)
	}

	// Test JSONRPCRequest, JSONRPCResponse, JSONRPCNotification, JSONRPCError
	req := ClientRequest{
		JSONRPCRequest: JSONRPCRequest{
			JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
			ID:             NewIntRequestID(1),
			Method:         "test",
			Params:         json.RawMessage(`{"a":1}`),
		},
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid jsonrpc 2.0, got %v", err)
	}
	badMsg := JSONRPCMessage{JSONRPC: "1.0"}
	if err := badMsg.Validate(); err == nil {
		t.Errorf("expected error for jsonrpc 1.0")
	}
	b, err = json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"method":"test"`) {
		t.Fatalf("expected method in request: %v", err)
	}
	var reqOut ClientRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || reqOut.ID.String() != "1" || reqOut.Method != "test" {
		t.Errorf("ClientRequest round-trip failed: %+v, %v", reqOut, err)
	}

	res := ServerResult{
		JSONRPCResponse: JSONRPCResponse{
			JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
			ID:             NewIntRequestID(1),
			Result:         json.RawMessage(`"ok"`),
			Error:          &JSONRPCError{Code: 100, Message: "err", Data: json.RawMessage(`"details"`)},
		},
	}
	b, err = json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"message":"err"`) {
		t.Fatalf("expected error message in response: %v", err)
	}
	var resOut ServerResult
	if err := json.Unmarshal(b, &resOut); err != nil || resOut.Error == nil || resOut.Error.Code != 100 {
		t.Errorf("ServerResult round-trip failed: %+v, %v", resOut, err)
	}

	not := ClientNotification{
		JSONRPCNotification: JSONRPCNotification{
			JSONRPCMessage: JSONRPCMessage{JSONRPC: "2.0"},
			Method:         "notify",
		},
	}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"method":"notify"`) {
		t.Fatalf("expected method in notification: %v", err)
	}

	baseRes := Result{Meta: &Meta{ProgressToken: &intTok}}
	b, err = json.Marshal(baseRes)
	if err != nil || !strings.Contains(string(b), `"_meta"`) {
		t.Fatalf("expected _meta: %v", err)
	}

	// Test RequestID methods and UnmarshalRequestID
	var emptyID RequestID
	if !emptyID.IsZero() || emptyID.String() != "" {
		t.Errorf("expected zero RequestID")
	}
	nullIDBytes, _ := json.Marshal(emptyID)
	if string(nullIDBytes) != "null" {
		t.Errorf("expected null for empty RequestID")
	}
	if err := json.Unmarshal([]byte(`null`), &emptyID); err != nil || !emptyID.IsZero() {
		t.Errorf("expected zero RequestID on null")
	}

	idStr, err := UnmarshalRequestID([]byte(`"req-1"`))
	if err != nil || idStr.String() != "req-1" {
		t.Errorf("expected req-1, got %v, %v", idStr, err)
	}
	idInt, err := UnmarshalRequestID([]byte(`42`))
	if err != nil || idInt.String() != "42" {
		t.Errorf("expected 42, got %v, %v", idInt, err)
	}
	if _, err := UnmarshalRequestID([]byte(`null`)); err == nil {
		t.Errorf("expected error for null in UnmarshalRequestID")
	}
	if _, err := UnmarshalRequestID([]byte(`true`)); err == nil {
		t.Errorf("expected error for boolean RequestID")
	}
}
