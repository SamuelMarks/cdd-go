package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaSamplingStructs(t *testing.T) {
	// Test ModelPreferences
	prefs := ModelPreferences{CostPriority: 1.0, Hints: []ModelHint{{Name: "claude-3-5-sonnet"}}}
	b, _ := json.Marshal(prefs)
	if !strings.Contains(string(b), `"claude-3-5-sonnet"`) || !strings.Contains(string(b), `"costPriority":1`) {
		t.Errorf("expected model preferences fields")
	}

	// Test FullCreateMessageRequest
	req := FullCreateMessageRequest{Method: "sampling/createMessage"}
	req.Params.Messages = []SamplingMessage{{Role: "user", Content: "Hello"}}
	req.Params.MaxTokens = 100
	req.Params.ModelPreferences = &prefs
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"sampling/createMessage"`) || !strings.Contains(string(b), `"user"`) || !strings.Contains(string(b), `"maxTokens":100`) {
		t.Errorf("expected create message request fields")
	}

	// Test CreateMessageResult
	res := CreateMessageResult{Role: RoleAssistant, Model: "claude-3-5-sonnet-20241022"}
	res.Content = TextContent{Type: "text", Text: "Hi"}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"assistant"`) || !strings.Contains(string(b), `"claude-3-5-sonnet-20241022"`) || !strings.Contains(string(b), `"Hi"`) {
		t.Errorf("expected create message result fields")
	}
}
