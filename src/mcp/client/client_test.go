package client

import (
	"context"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// mockTransport implements mcp.Transport for testing.
type mockTransport struct{}

func (m *mockTransport) Send(msg []byte) error {
	return nil
}

func (m *mockTransport) Receive() ([]byte, error) {
	return nil, nil
}

func (m *mockTransport) Close() error {
	return nil
}

func TestMCPClient(t *testing.T) {
	c := NewMCPClient(&mockTransport{})

	ctx := context.Background()

	tools, err := c.GetTools(ctx)
	if err != nil {
		t.Errorf("GetTools failed: %v", err)
	}
	if tools == nil || tools.Tools == nil {
		t.Errorf("expected empty tools slice")
	}

	res, err := c.ExecuteTool(ctx, "test_tool", map[string]interface{}{"foo": "bar"})
	if err != nil {
		t.Errorf("ExecuteTool failed: %v", err)
	}
	if len(res.Content) == 0 {
		t.Errorf("expected content")
	}

	resources, err := c.GetResources(ctx)
	if err != nil {
		t.Errorf("GetResources failed: %v", err)
	}
	if resources == nil || resources.Resources == nil {
		t.Errorf("expected empty resources slice")
	}

	if tc, ok := res.Content[0].(mcp.TextContent); ok {
		if !strings.Contains(tc.Text, "Executed test_tool") {
			t.Errorf("expected executed text")
		}
	} else {
		t.Errorf("expected TextContent")
	}
}

func TestMCPClient_GetToolsPaginated(t *testing.T) {
	c := NewMCPClient(nil)
	res, err := c.GetToolsPaginated(context.Background(), "cursor1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NextCursor != "cursor1" {
		t.Errorf("expected cursor1, got %v", res.NextCursor)
	}
}

func TestMCPClient_GetResourcesPaginated(t *testing.T) {
	c := NewMCPClient(nil)
	res, err := c.GetResourcesPaginated(context.Background(), "cursor2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.NextCursor != "cursor2" {
		t.Errorf("expected cursor2, got %v", res.NextCursor)
	}
}

func TestMCPClient_SendProgress(t *testing.T) {
	c := NewMCPClient(nil)
	err := c.SendProgress(context.Background(), "tok", 50, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMCPClient_CreateMessage(t *testing.T) {
	c := NewMCPClient(nil)
	res, err := c.CreateMessage(context.Background(), []mcp.SamplingMessage{{Role: "user", Content: "hi"}}, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Role != "assistant" {
		t.Errorf("expected assistant role")
	}
}

func TestMCPClient_ResolveURI(t *testing.T) {
	c := NewMCPClient(nil)
	res, err := c.ResolveURI(context.Background(), "custom://foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res) != "Resolved custom://foo" {
		t.Errorf("expected resolution")
	}
}

func TestMCPClient_IsPathAllowed(t *testing.T) {
	c := NewMCPClient(nil)
	if !c.IsPathAllowed(context.Background(), "/tmp") {
		t.Errorf("expected true")
	}
}

func TestMCPClient_InspectSchema(t *testing.T) {
	c := NewMCPClient(nil)
	res, err := c.InspectSchema(context.Background(), "openapi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "{}" {
		t.Errorf("expected {}")
	}
}
