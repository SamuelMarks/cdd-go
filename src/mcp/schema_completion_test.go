package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaCompletionStructs(t *testing.T) {
	// Test CompleteRequest
	req := CompleteRequest{Method: "completion/complete"}
	req.Params.Ref = PromptReference{Type: "ref/prompt", Name: "my_prompt"}
	req.Params.Argument.Name = "arg1"
	req.Params.Argument.Value = "val"
	b, _ := json.Marshal(req)
	if !strings.Contains(string(b), `"completion/complete"`) || !strings.Contains(string(b), `"arg1"`) || !strings.Contains(string(b), `"val"`) {
		t.Errorf("expected complete request fields")
	}

	// Test CompleteResult
	res := CompleteResult{}
	res.Completion.Values = []string{"val1", "val2"}
	res.Completion.Total = 2
	res.Completion.HasMore = false
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"val1"`) || !strings.Contains(string(b), `"total":2`) {
		t.Errorf("expected complete result fields")
	}
}
