package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaCompletionStructsAndRoundTrip(t *testing.T) {
	req := CompleteRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "completion/complete"}
	req.Params.Ref = CompletionReference{PromptRef: &PromptReference{Type: "ref/prompt", Name: "p1"}}
	req.Params.Argument.Name = "arg"
	req.Params.Argument.Value = "val"
	b, err := json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"completion/complete"`) || !strings.Contains(string(b), `"val"`) {
		t.Fatalf("expected complete request fields: %v", err)
	}
	var reqOut CompleteRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || !reflect.DeepEqual(req, reqOut) {
		t.Errorf("CompleteRequest promptRef round-trip mismatch: %+v vs %+v (%v)", req, reqOut, err)
	}

	// Test ResourceReference branch in CompletionReference
	resRef := CompletionReference{ResourceRef: &ResourceReference{Type: "ref/resource", URI: "file://test"}}
	bRef, err := json.Marshal(resRef)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var resRefOut CompletionReference
	if err := json.Unmarshal(bRef, &resRefOut); err != nil || resRefOut.ResourceRef == nil || resRefOut.ResourceRef.URI != "file://test" {
		t.Errorf("CompletionReference resourceRef round-trip failed: %+v, %v", resRefOut, err)
	}

	// Test empty and invalid CompletionReference
	emptyBytes, _ := json.Marshal(CompletionReference{})
	if string(emptyBytes) != "null" {
		t.Errorf("expected null for empty CompletionReference")
	}
	if err := json.Unmarshal([]byte(`123`), &resRefOut); err == nil {
		t.Errorf("expected error for non-object CompletionReference JSON")
	}
	if err := json.Unmarshal([]byte(`{"type":"ref/prompt","name":123}`), &resRefOut); err == nil {
		t.Errorf("expected error for invalid prompt ref")
	}
	if err := json.Unmarshal([]byte(`{"type":"ref/resource","uri":123}`), &resRefOut); err == nil {
		t.Errorf("expected error for invalid resource ref")
	}
	if err := json.Unmarshal([]byte(`{"type":"unknown"}`), &resRefOut); err == nil {
		t.Errorf("expected error for unknown ref type")
	}

	res := CompleteResult{}
	res.Completion.Values = []string{"val1", "val2"}
	res.Completion.Total = 2
	res.Completion.HasMore = false
	b, err = json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"val1"`) || !strings.Contains(string(b), `"total":2`) {
		t.Fatalf("expected complete result fields: %v", err)
	}
	var resOut CompleteResult
	if err := json.Unmarshal(b, &resOut); err != nil || !reflect.DeepEqual(res, resOut) {
		t.Errorf("CompleteResult round-trip mismatch: %+v vs %+v (%v)", res, resOut, err)
	}
}
