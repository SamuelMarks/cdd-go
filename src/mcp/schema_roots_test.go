package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaRootsStructs(t *testing.T) {
	// Test Root
	root := Root{URI: "file://root", Name: "RootName"}
	b, _ := json.Marshal(root)
	if !strings.Contains(string(b), `"file://root"`) || !strings.Contains(string(b), `"RootName"`) {
		t.Errorf("expected uri and name")
	}

	// Test ListRootsRequest
	req := ListRootsRequest{Method: "roots/list"}
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"roots/list"`) {
		t.Errorf("expected method")
	}

	// Test ListRootsResult
	res := ListRootsResult{Roots: []Root{root}}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"RootName"`) {
		t.Errorf("expected root in result")
	}

	// Test RootsListChangedNotification
	not := RootsListChangedNotification{Method: "notifications/roots/list_changed"}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"notifications/roots/list_changed"`) {
		t.Errorf("expected method")
	}
}
