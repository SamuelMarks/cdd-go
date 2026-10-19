package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCapabilitiesStructsAndRoundTrip(t *testing.T) {
	clientCaps := ClientCapabilities{
		Experimental: map[string]json.RawMessage{"featureX": json.RawMessage(`true`)},
		Roots: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{ListChanged: true},
		Sampling: &struct{}{},
	}
	b, err := json.Marshal(clientCaps)
	if err != nil || !strings.Contains(string(b), `"listChanged":true`) || !strings.Contains(string(b), `"sampling":{}`) {
		t.Fatalf("expected roots and sampling in client caps: %v", err)
	}
	var clientCapsOut ClientCapabilities
	if err := json.Unmarshal(b, &clientCapsOut); err != nil || !reflect.DeepEqual(clientCaps, clientCapsOut) {
		t.Errorf("ClientCapabilities round-trip mismatch: %+v vs %+v (%v)", clientCaps, clientCapsOut, err)
	}

	serverCaps := ServerCapabilities{
		Logging: &struct{}{},
		Prompts: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{ListChanged: true},
		Resources: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
			Subscribe   bool `json:"subscribe,omitempty"`
		}{Subscribe: true},
		Tools: &struct {
			ListChanged bool `json:"listChanged,omitempty"`
		}{ListChanged: true},
	}
	b, err = json.Marshal(serverCaps)
	if err != nil || !strings.Contains(string(b), `"subscribe":true`) || !strings.Contains(string(b), `"tools"`) {
		t.Fatalf("expected subscribe and tools in server caps: %v", err)
	}
	var serverCapsOut ServerCapabilities
	if err := json.Unmarshal(b, &serverCapsOut); err != nil || !reflect.DeepEqual(serverCaps, serverCapsOut) {
		t.Errorf("ServerCapabilities round-trip mismatch: %+v vs %+v (%v)", serverCaps, serverCapsOut, err)
	}
}
