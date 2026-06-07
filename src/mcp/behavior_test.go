package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBehaviorStructs(t *testing.T) {
	// Test ProgressNotification
	prog := ProgressNotification{JSONRPC: "2.0", Method: "$/progress"}
	prog.Params.ProgressToken = "task-1"
	prog.Params.Progress = 50.0
	prog.Params.Total = 100.0
	b, _ := json.Marshal(prog)
	if !strings.Contains(string(b), `"progress":50`) {
		t.Errorf("expected progress")
	}

	// Test PaginatedRequest
	reqPage := PaginatedRequest{Method: "list/things"}
	reqPage.Params.Cursor = "cursor1"
	b, _ = json.Marshal(reqPage)
	if !strings.Contains(string(b), `"cursor1"`) {
		t.Errorf("expected cursor")
	}

	// Test PaginatedResult
	page := PaginatedResult{
		Meta:       &Meta{ProgressToken: "test"},
		NextCursor: "cursor123",
	}
	b, _ = json.Marshal(page)
	if !strings.Contains(string(b), `"nextCursor":"cursor123"`) {
		t.Errorf("expected cursor")
	}
	if !strings.Contains(string(b), `"progressToken":"test"`) {
		t.Errorf("expected meta progressToken")
	}
}

func TestSecurityManager(t *testing.T) {
	sm := &SecurityManager{
		AllowedRoots:    []string{"/safe/dir", "/var/log"},
		RequireApproval: true,
	}

	if !sm.IsPathAllowed("/safe/dir/file.txt") {
		t.Errorf("expected path to be allowed")
	}
	if sm.IsPathAllowed("/etc/passwd") {
		t.Errorf("expected path to be denied")
	}

	if !sm.RequiresToolApproval("rm") {
		t.Errorf("expected tool to require approval")
	}

	// Secure by default test
	smEmpty := &SecurityManager{}
	if smEmpty.IsPathAllowed("/anything") {
		t.Errorf("expected empty roots to deny all")
	}
}

func TestSamplingAndProtocols(t *testing.T) {
	// Test CreateMessageRequest
	req := CreateMessageRequest{JSONRPC: "2.0", Method: "sampling/createMessage"}
	req.Params.Messages = []SamplingMessage{{Role: "user", Content: "Hello"}}
	b, _ := json.Marshal(req)
	if !strings.Contains(string(b), `"role":"user"`) {
		t.Errorf("expected user role")
	}

	// Test ProtocolHandler
	ph := &ProtocolHandler{SupportedSchemes: []string{"file://", "https://"}}
	if !ph.IsSchemeSupported("file:///tmp/test.txt") {
		t.Errorf("expected file to be supported")
	}
	if ph.IsSchemeSupported("ftp://server") {
		t.Errorf("expected ftp to not be supported")
	}
}
