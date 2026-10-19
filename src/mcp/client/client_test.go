package client

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// mockTransport implements mcp.Transport for testing.
type mockTransport struct {
	sent      [][]byte
	responses [][]byte
	sendErr   error
	recvErr   error
}

func (m *mockTransport) Send(msg []byte) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, msg)
	return nil
}

func (m *mockTransport) Receive() ([]byte, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	if len(m.responses) > 0 {
		r := m.responses[0]
		m.responses = m.responses[1:]
		return r, nil
	}
	return nil, io.EOF
}

func (m *mockTransport) Close() error { return nil }

func TestMCPClient(t *testing.T) {
	mt := &mockTransport{}
	c := NewMCPClient(mt)
	ctx := context.Background()

	// Fallback when transport returns EOF
	tools, err := c.GetTools(ctx)
	if err != nil || tools == nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Real JSON-RPC response from transport
	mt.responses = [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"gen","inputSchema":{"type":"object"}}],"nextCursor":"c2"}}`),
	}
	tools, err = c.GetToolsPaginated(ctx, "c1")
	if err != nil || len(tools.Tools) != 1 || tools.Tools[0].Name != "gen" || tools.NextCursor != "c2" {
		t.Errorf("expected decoded tools from transport, got %+v, %v", tools, err)
	}

	// ExecuteTool with fallback and with real response
	res, err := c.ExecuteTool(ctx, "test_tool", map[string]string{"foo": "bar"})
	if err != nil || len(res.Content) != 1 {
		t.Fatalf("unexpected error: %v", err)
	}
	mt.responses = [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"real tool output"}]}}`),
	}
	res, err = c.ExecuteTool(ctx, "test_tool", map[string]string{"foo": "bar"})
	if err != nil || res.Content[0].(mcp.TextContent).Text != "real tool output" {
		t.Errorf("expected real tool output, got %+v, %v", res, err)
	}

	// GetResources with fallback and with real response
	resources, err := c.GetResources(ctx)
	if err != nil || resources == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mt.responses = [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":6,"result":{"resources":[{"uri":"file://a","name":"A"}]}}`),
	}
	resources, err = c.GetResourcesPaginated(ctx, "c1")
	if err != nil || len(resources.Resources) != 1 || resources.Resources[0].Name != "A" {
		t.Errorf("expected decoded resources, got %+v, %v", resources, err)
	}

	// SendProgress
	if err := c.SendProgress(ctx, mcp.NewStringProgressToken("tok1"), 50, 100); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nilClient := NewMCPClient(nil)
	if err := nilClient.SendProgress(ctx, mcp.NewIntProgressToken(1), 10, 20); err != nil {
		t.Fatalf("unexpected error on nil transport: %v", err)
	}
	if _, err := nilClient.GetTools(ctx); err != nil {
		t.Fatalf("unexpected error on nil transport GetTools: %v", err)
	}

	// CreateMessage with fallback and with real response
	msgRes, err := c.CreateMessage(ctx, []mcp.SamplingMessage{{Role: "user", Content: "hi"}}, 100)
	if err != nil || msgRes.Role != mcp.RoleAssistant {
		t.Fatalf("unexpected error: %v", err)
	}
	mt.responses = [][]byte{
		[]byte(`{"jsonrpc":"2.0","id":8,"result":{"role":"assistant","model":"m1","content":{"type":"text","text":"sampled!"}}}`),
	}
	msgRes, err = c.CreateMessage(ctx, []mcp.SamplingMessage{{Role: "user", Content: "hi"}}, 100)
	if err != nil || msgRes.Model != "m1" {
		t.Errorf("expected decoded CreateMessageResult, got %+v, %v", msgRes, err)
	}

	// ResolveURI
	data, err := c.ResolveURI(ctx, "mcp://resource")
	if err != nil || !strings.Contains(string(data), "Resolved mcp://resource") {
		t.Fatalf("unexpected ResolveURI result: %s, %v", string(data), err)
	}
	if _, err := c.ResolveURI(ctx, "   "); err == nil {
		t.Errorf("expected error for empty uri")
	}

	// IsPathAllowed
	if !c.IsPathAllowed(ctx, "/valid/path") {
		t.Errorf("expected path to be allowed")
	}

	// InspectSchema
	schemaJSON, err := c.InspectSchema(ctx, "openapi")
	if err != nil || schemaJSON != `{"schemaType":"openapi"}` {
		t.Errorf("unexpected InspectSchema result: %q, %v", schemaJSON, err)
	}
	if _, err := c.InspectSchema(ctx, ""); err == nil {
		t.Errorf("expected error for empty schemaType")
	}
}

func TestMCPClientErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewMCPClient(&mockTransport{})
	if _, err := c.GetTools(ctx); err == nil {
		t.Error("expected context canceled error on GetTools")
	}
	if _, err := c.ExecuteTool(ctx, "t", nil); err == nil {
		t.Error("expected context canceled error on ExecuteTool")
	}
	if _, err := c.GetResources(ctx); err == nil {
		t.Error("expected context canceled error on GetResources")
	}
	if err := c.SendProgress(ctx, mcp.NewIntProgressToken(1), 1, 2); err == nil {
		t.Error("expected context canceled error on SendProgress")
	}
	if _, err := c.CreateMessage(ctx, nil, 10); err == nil {
		t.Error("expected context canceled error on CreateMessage")
	}
	if _, err := c.ResolveURI(ctx, "mcp://a"); err == nil {
		t.Error("expected context canceled error on ResolveURI")
	}
	if _, err := c.InspectSchema(ctx, "openapi"); err == nil {
		t.Error("expected context canceled error on InspectSchema")
	}

	// Transport Send error
	bg := context.Background()
	sendErrClient := NewMCPClient(&mockTransport{sendErr: errors.New("send failed")})
	if _, err := sendErrClient.GetTools(bg); err == nil {
		t.Error("expected send error")
	}

	// Transport Receive error
	recvErrClient := NewMCPClient(&mockTransport{recvErr: errors.New("recv failed")})
	if _, err := recvErrClient.GetTools(bg); err == nil {
		t.Error("expected recv error")
	}

	// Empty response bytes, RPC error response, and malformed result JSON
	mt := &mockTransport{
		responses: [][]byte{
			[]byte(``),
			[]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`),
			[]byte(`{"jsonrpc":"2.0","id":2,"result":"not-an-object"}`),
			[]byte(`not-json`),
		},
	}
	errClient := NewMCPClient(mt)
	if _, err := errClient.GetTools(bg); err != nil {
		t.Errorf("expected fallback on empty response bytes, got %v", err)
	}
	if _, err := errClient.GetTools(bg); err == nil {
		t.Error("expected RPC error from response")
	}
	if _, err := errClient.GetTools(bg); err == nil {
		t.Error("expected unmarshal error on invalid result type")
	}
	if _, err := errClient.GetTools(bg); err != nil {
		t.Errorf("expected fallback on non-JSON response, got %v", err)
	}
}
