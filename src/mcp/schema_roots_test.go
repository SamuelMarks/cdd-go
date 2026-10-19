package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaRootsStructsAndRoundTrip(t *testing.T) {
	root := Root{URI: "file://root", Name: "RootName"}
	b, err := json.Marshal(root)
	if err != nil || !strings.Contains(string(b), `"file://root"`) || !strings.Contains(string(b), `"RootName"`) {
		t.Fatalf("expected root fields: %v", err)
	}
	var rootOut Root
	if err := json.Unmarshal(b, &rootOut); err != nil || !reflect.DeepEqual(root, rootOut) {
		t.Errorf("Root round-trip mismatch: %+v vs %+v (%v)", root, rootOut, err)
	}

	req := ListRootsRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "roots/list"}
	b, err = json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"roots/list"`) {
		t.Fatalf("expected method: %v", err)
	}
	var reqOut ListRootsRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || !reflect.DeepEqual(req, reqOut) {
		t.Errorf("ListRootsRequest round-trip mismatch: %+v vs %+v (%v)", req, reqOut, err)
	}

	res := ListRootsResult{Roots: []Root{root}}
	b, err = json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"RootName"`) {
		t.Fatalf("expected roots in result: %v", err)
	}
	var resOut ListRootsResult
	if err := json.Unmarshal(b, &resOut); err != nil || !reflect.DeepEqual(res, resOut) {
		t.Errorf("ListRootsResult round-trip mismatch: %+v vs %+v (%v)", res, resOut, err)
	}

	not := RootsListChangedNotification{JSONRPC: "2.0", Method: "notifications/roots/list_changed"}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"notifications/roots/list_changed"`) {
		t.Fatalf("expected method: %v", err)
	}
}
