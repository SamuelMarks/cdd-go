package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestBehaviorStructsAndRoundTrip(t *testing.T) {
	prog := ProgressNotification{JSONRPC: "2.0", Method: "$/progress"}
	prog.Params.ProgressToken = NewStringProgressToken("tok1")
	prog.Params.Progress = 50.0
	prog.Params.Total = 100.0
	b, err := json.Marshal(prog)
	if err != nil || !strings.Contains(string(b), `"progress":50`) {
		t.Fatalf("expected progress 50: %v", err)
	}
	var progOut ProgressNotification
	if err := json.Unmarshal(b, &progOut); err != nil || !reflect.DeepEqual(prog, progOut) {
		t.Errorf("ProgressNotification round-trip mismatch: %+v vs %+v (%v)", prog, progOut, err)
	}

	pagReq := PaginatedRequest{Method: "resources/list"}
	pagReq.Params.Cursor = "cur1"
	b, err = json.Marshal(pagReq)
	if err != nil || !strings.Contains(string(b), `"cursor":"cur1"`) {
		t.Fatalf("expected cursor cur1: %v", err)
	}

	pagRes := PaginatedResult{NextCursor: "cur2", Items: []json.RawMessage{json.RawMessage(`"a"`), json.RawMessage(`"b"`)}}
	b, err = json.Marshal(pagRes)
	if err != nil || !strings.Contains(string(b), `"nextCursor":"cur2"`) {
		t.Fatalf("expected nextCursor cur2: %v", err)
	}
	var pagResOut PaginatedResult
	if err := json.Unmarshal(b, &pagResOut); err != nil || !reflect.DeepEqual(pagRes, pagResOut) {
		t.Errorf("PaginatedResult round-trip mismatch: %+v vs %+v (%v)", pagRes, pagResOut, err)
	}

	createMsg := CreateMessageRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "sampling/createMessage"}
	createMsg.Params.Messages = []SamplingMessage{{Role: "user", Content: "hello"}}
	createMsg.Params.SystemPrompt = "sys"
	createMsg.Params.MaxTokens = 100
	b, err = json.Marshal(createMsg)
	if err != nil || !strings.Contains(string(b), `"hello"`) {
		t.Fatalf("expected hello in message: %v", err)
	}
	var createMsgOut CreateMessageRequest
	if err := json.Unmarshal(b, &createMsgOut); err != nil || !reflect.DeepEqual(createMsg, createMsgOut) {
		t.Errorf("CreateMessageRequest round-trip mismatch: %+v vs %+v (%v)", createMsg, createMsgOut, err)
	}
}

func TestSecurityManager(t *testing.T) {
	sm := SecurityManager{AllowedRoots: []string{"/workspace", "/tmp"}, RequireApproval: true}
	if !sm.IsPathAllowed("/workspace/file.go") {
		t.Errorf("expected path to be allowed")
	}
	if sm.IsPathAllowed("/etc/passwd") {
		t.Errorf("expected path to be denied")
	}
	if !sm.RequiresToolApproval("any_tool") {
		t.Errorf("expected approval required")
	}

	smEmpty := SecurityManager{}
	if smEmpty.IsPathAllowed("/anything") {
		t.Errorf("expected empty roots to deny all")
	}
}

func TestProtocolHandler(t *testing.T) {
	ph := ProtocolHandler{SupportedSchemes: []string{"file://", "https://"}}
	if !ph.IsSchemeSupported("file:///tmp/test.txt") {
		t.Errorf("expected file to be supported")
	}
	if ph.IsSchemeSupported("ftp://server") {
		t.Errorf("expected ftp to not be supported")
	}
}
