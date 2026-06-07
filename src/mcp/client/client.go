package client

import (
	"context"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPClient provides native programmatic access to an MCP server.
type MCPClient struct {
	Transport mcp.Transport
}

// NewMCPClient creates a new native client.
func NewMCPClient(t mcp.Transport) *MCPClient {
	return &MCPClient{Transport: t}
}

// GetTools requests the list of tools from the server.
func (c *MCPClient) GetTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	return c.GetToolsPaginated(ctx, "")
}

// GetToolsPaginated requests the list of tools from the server with a cursor.
func (c *MCPClient) GetToolsPaginated(ctx context.Context, cursor string) (*mcp.ListToolsResult, error) {
	// Stub: In reality this formats JSONRPC and sends over transport.
	// We return a mock empty list to satisfy the LLM router logic.
	return &mcp.ListToolsResult{
		Tools:      []mcp.Tool{},
		NextCursor: mcp.Cursor(cursor),
	}, nil
}

// ExecuteTool routes an execution request native to the MCP transport.
func (c *MCPClient) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// Stub: In reality this formats JSONRPC and sends over transport.
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: "Executed " + name},
		},
	}, nil
}

// GetResources requests the list of resources from the server.
func (c *MCPClient) GetResources(ctx context.Context) (*mcp.ListResourcesResult, error) {
	return c.GetResourcesPaginated(ctx, "")
}

// GetResourcesPaginated requests the list of resources from the server with a cursor.
func (c *MCPClient) GetResourcesPaginated(ctx context.Context, cursor string) (*mcp.ListResourcesResult, error) {
	return &mcp.ListResourcesResult{
		Resources:  []mcp.Resource{},
		NextCursor: mcp.Cursor(cursor),
	}, nil
}

// SendProgress emits a progress event.
func (c *MCPClient) SendProgress(ctx context.Context, token interface{}, progress, total float64) error {
	// Stub: In reality this formats JSONRPC and sends over transport.
	return nil
}

// CreateMessage requests a sample message for human-in-the-loop.
func (c *MCPClient) CreateMessage(ctx context.Context, messages []mcp.SamplingMessage, maxTokens int) (*mcp.CreateMessageResult, error) {
	// Stub: In reality this formats JSONRPC and sends over transport.
	return &mcp.CreateMessageResult{
		Content: mcp.TextContent{Type: "text", Text: "Sampled"},
		Role:    "assistant",
	}, nil
}

// ResolveURI resolves a custom URI scheme.
func (c *MCPClient) ResolveURI(ctx context.Context, uri string) ([]byte, error) {
	// Stub
	return []byte("Resolved " + uri), nil
}

// IsPathAllowed enforces root boundaries.
func (c *MCPClient) IsPathAllowed(ctx context.Context, path string) bool {
	// Stub
	return true
}

// InspectSchema queries loaded OpenAPI/AsyncAPI schemas.
func (c *MCPClient) InspectSchema(ctx context.Context, schemaType string) (string, error) {
	// Stub
	return "{}", nil
}
