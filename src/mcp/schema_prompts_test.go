package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaPromptsStructsAndRoundTrip(t *testing.T) {
	arg := PromptArgument{Name: "arg1", Description: "first arg", Required: true}
	b, err := json.Marshal(arg)
	if err != nil || !strings.Contains(string(b), `"name":"arg1"`) || !strings.Contains(string(b), `"required":true`) {
		t.Fatalf("expected prompt argument fields: %v", err)
	}

	prompt := Prompt{Name: "p1", Description: "desc", Arguments: []PromptArgument{arg}}
	b, err = json.Marshal(prompt)
	if err != nil || !strings.Contains(string(b), `"name":"p1"`) {
		t.Fatalf("expected prompt name: %v", err)
	}
	var promptOut Prompt
	if err := json.Unmarshal(b, &promptOut); err != nil || !reflect.DeepEqual(prompt, promptOut) {
		t.Errorf("Prompt round-trip mismatch: %+v vs %+v (%v)", prompt, promptOut, err)
	}

	msg := PromptMessage{Role: RoleUser, Content: TextContent{Type: "text", Text: "hello"}}
	b, err = json.Marshal(msg)
	if err != nil || !strings.Contains(string(b), `"role":"user"`) {
		t.Fatalf("expected prompt message role: %v", err)
	}
	var msgOut PromptMessage
	if err := json.Unmarshal(b, &msgOut); err != nil || !reflect.DeepEqual(msg, msgOut) {
		t.Errorf("PromptMessage round-trip mismatch: %+v vs %+v (%v)", msg, msgOut, err)
	}
	if err := json.Unmarshal([]byte(`123`), &msgOut); err == nil {
		t.Errorf("expected error for non-object PromptMessage json")
	}
	if err := json.Unmarshal([]byte(`{"role":"user","content":{"type":"bad"}}`), &msgOut); err == nil {
		t.Errorf("expected error for invalid PromptMessage content")
	}

	ref := PromptReference{Type: "ref/prompt", Name: "p1"}
	b, err = json.Marshal(ref)
	if err != nil || !strings.Contains(string(b), `"ref/prompt"`) {
		t.Fatalf("expected prompt reference: %v", err)
	}

	getReq := GetPromptRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "prompts/get"}
	getReq.Params.Name = "p1"
	getReq.Params.Arguments = map[string]string{"k": "v"}
	b, err = json.Marshal(getReq)
	if err != nil || !strings.Contains(string(b), `"prompts/get"`) || !strings.Contains(string(b), `"k":"v"`) {
		t.Fatalf("expected get prompt request fields: %v", err)
	}
	var getReqOut GetPromptRequest
	if err := json.Unmarshal(b, &getReqOut); err != nil || getReqOut.Params.Arguments["k"] != "v" {
		t.Errorf("GetPromptRequest round-trip failed: %+v, %v", getReqOut, err)
	}

	getRes := GetPromptResult{Description: "res", Messages: []PromptMessage{msg}}
	b, err = json.Marshal(getRes)
	if err != nil || !strings.Contains(string(b), `"description":"res"`) {
		t.Fatalf("expected get prompt result description: %v", err)
	}
	var getResOut GetPromptResult
	if err := json.Unmarshal(b, &getResOut); err != nil || !reflect.DeepEqual(getRes, getResOut) {
		t.Errorf("GetPromptResult round-trip mismatch: %+v vs %+v (%v)", getRes, getResOut, err)
	}

	listReq := ListPromptsRequest{JSONRPC: "2.0", ID: NewIntRequestID(2), Method: "prompts/list"}
	listReq.Params.Cursor = "c1"
	b, err = json.Marshal(listReq)
	if err != nil || !strings.Contains(string(b), `"prompts/list"`) || !strings.Contains(string(b), `"c1"`) {
		t.Fatalf("expected list prompts request: %v", err)
	}

	listRes := ListPromptsResult{NextCursor: "c2", Prompts: []Prompt{prompt}}
	b, err = json.Marshal(listRes)
	if err != nil || !strings.Contains(string(b), `"c2"`) {
		t.Fatalf("expected list prompts result: %v", err)
	}
	var listResOut ListPromptsResult
	if err := json.Unmarshal(b, &listResOut); err != nil || !reflect.DeepEqual(listRes, listResOut) {
		t.Errorf("ListPromptsResult round-trip mismatch: %+v vs %+v (%v)", listRes, listResOut, err)
	}

	not := PromptListChangedNotification{JSONRPC: "2.0", Method: "notifications/prompts/list_changed"}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"notifications/prompts/list_changed"`) {
		t.Fatalf("expected notification method: %v", err)
	}
}
