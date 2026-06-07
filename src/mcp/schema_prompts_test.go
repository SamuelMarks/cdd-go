package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaPromptsStructs(t *testing.T) {
	// Test PromptArgument
	arg := PromptArgument{Name: "arg1", Required: true}
	b, _ := json.Marshal(arg)
	if !strings.Contains(string(b), `"arg1"`) || !strings.Contains(string(b), `"required":true`) {
		t.Errorf("expected prompt argument fields")
	}

	// Test Prompt
	prompt := Prompt{Name: "p1", Arguments: []PromptArgument{arg}}
	b, _ = json.Marshal(prompt)
	if !strings.Contains(string(b), `"p1"`) || !strings.Contains(string(b), `"arg1"`) {
		t.Errorf("expected prompt fields")
	}

	// Test PromptMessage
	msg := PromptMessage{Role: RoleUser, Content: TextContent{Type: "text", Text: "hi"}}
	b, _ = json.Marshal(msg)
	if !strings.Contains(string(b), `"user"`) || !strings.Contains(string(b), `"hi"`) {
		t.Errorf("expected prompt message fields")
	}

	// Test PromptReference
	ref := PromptReference{Type: "ref/prompt", Name: "p1"}
	b, _ = json.Marshal(ref)
	if !strings.Contains(string(b), `"ref/prompt"`) {
		t.Errorf("expected ref/prompt")
	}

	// Test GetPromptRequest
	req := GetPromptRequest{Method: "prompts/get"}
	req.Params.Name = "p1"
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"prompts/get"`) || !strings.Contains(string(b), `"p1"`) {
		t.Errorf("expected get prompt fields")
	}

	// Test GetPromptResult
	res := GetPromptResult{Messages: []PromptMessage{msg}}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"hi"`) {
		t.Errorf("expected messages in result")
	}

	// Test ListPromptsRequest
	listReq := ListPromptsRequest{Method: "prompts/list"}
	b, _ = json.Marshal(listReq)
	if !strings.Contains(string(b), `"prompts/list"`) {
		t.Errorf("expected list method")
	}

	// Test ListPromptsResult
	listRes := ListPromptsResult{Prompts: []Prompt{prompt}}
	b, _ = json.Marshal(listRes)
	if !strings.Contains(string(b), `"p1"`) {
		t.Errorf("expected prompts in result")
	}

	// Test PromptListChangedNotification
	not := PromptListChangedNotification{Method: "notifications/prompts/list_changed"}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"notifications/prompts/list_changed"`) {
		t.Errorf("expected notification method")
	}
}
