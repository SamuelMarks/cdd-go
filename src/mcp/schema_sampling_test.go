package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaSamplingStructsAndRoundTrip(t *testing.T) {
	prefs := ModelPreferences{
		CostPriority:         1.0,
		SpeedPriority:        0.5,
		IntelligencePriority: 0.8,
		Hints:                []ModelHint{{Name: "claude"}},
	}
	b, err := json.Marshal(prefs)
	if err != nil || !strings.Contains(string(b), `"claude"`) || !strings.Contains(string(b), `"costPriority":1`) {
		t.Fatalf("expected model preferences fields: %v", err)
	}
	var prefsOut ModelPreferences
	if err := json.Unmarshal(b, &prefsOut); err != nil || !reflect.DeepEqual(prefs, prefsOut) {
		t.Errorf("ModelPreferences round-trip mismatch: %+v vs %+v (%v)", prefs, prefsOut, err)
	}

	req := FullCreateMessageRequest{
		JSONRPC: "2.0",
		ID:      NewIntRequestID(1),
		Method:  "sampling/createMessage",
		Params: CreateMessageRequestParams{
			MaxTokens:        100,
			SystemPrompt:     "sys",
			ModelPreferences: &prefs,
			Messages:         []SamplingMessage{{Role: "user", Content: "hello"}},
			Metadata:         map[string]json.RawMessage{"provider": json.RawMessage(`"anthropic"`)},
		},
	}
	b, err = json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"maxTokens":100`) || !strings.Contains(string(b), `"sampling/createMessage"`) {
		t.Fatalf("expected create message request fields: %v", err)
	}
	var reqOut FullCreateMessageRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || reqOut.Params.MaxTokens != 100 {
		t.Errorf("FullCreateMessageRequest round-trip failed: %+v, %v", reqOut, err)
	}

	res := CreateMessageResult{
		Role:       RoleAssistant,
		Model:      "claude-3-5-sonnet-20241022",
		StopReason: "endTurn",
		Content:    TextContent{Type: "text", Text: "Hi"},
	}
	b, err = json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"assistant"`) || !strings.Contains(string(b), `"claude-3-5-sonnet-20241022"`) || !strings.Contains(string(b), `"Hi"`) {
		t.Fatalf("expected create message result fields: %v", err)
	}
	var resOut CreateMessageResult
	if err := json.Unmarshal(b, &resOut); err != nil || !reflect.DeepEqual(res, resOut) {
		t.Errorf("CreateMessageResult round-trip mismatch: %+v vs %+v (%v)", res, resOut, err)
	}
	if err := json.Unmarshal([]byte(`123`), &resOut); err == nil {
		t.Errorf("expected error for non-object CreateMessageResult json")
	}
	if err := json.Unmarshal([]byte(`{"role":"assistant","content":{"type":"bad"}}`), &resOut); err == nil {
		t.Errorf("expected error for invalid CreateMessageResult content")
	}
}
