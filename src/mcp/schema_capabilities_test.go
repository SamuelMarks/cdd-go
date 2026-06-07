package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCapabilitiesStructs(t *testing.T) {
	// Test ClientCapabilities
	clientCaps := ClientCapabilities{
		Roots: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{ListChanged: true},
	}
	b, _ := json.Marshal(clientCaps)
	if !strings.Contains(string(b), `"listChanged":true`) {
		t.Errorf("expected listChanged in client roots")
	}

	// Test ServerCapabilities
	serverCaps := ServerCapabilities{
		Tools: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{ListChanged: false},
		Resources: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
			Subscribe   bool `json:"subscribe,omitempty"`
		}{Subscribe: true},
	}
	b, _ = json.Marshal(serverCaps)
	if !strings.Contains(string(b), `"subscribe":true`) || !strings.Contains(string(b), `"tools"`) {
		t.Errorf("expected subscribe and tools in server caps")
	}
}
