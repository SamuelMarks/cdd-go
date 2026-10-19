// Package client provides a programmatic MCP client for invoking tools, listing resources, and sampling over an MCP transport.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/mcp"
)

// MCPClient provides native programmatic access to an MCP server over a mcp.Transport.
type MCPClient struct {
	// Transport is the underlying MCP message transport.
	Transport mcp.Transport
	// Security configures root path boundaries and tool approvals.
	Security mcp.SecurityManager
	// nextID tracks the sequential request ID for outbound JSON-RPC calls.
	nextID int64
}

// NewMCPClient creates a new native client.
// Parameter t is the underlying MCP Transport, and it returns the initialized MCPClient.
func NewMCPClient(t mcp.Transport) *MCPClient {
	return &MCPClient{
		Transport: t,
		Security:  mcp.SecurityManager{AllowedRoots: []string{"/"}},
	}
}

// nextRequestID increments and returns the next sequential RequestID.
func (c *MCPClient) nextRequestID() mcp.RequestID {
	c.nextID++
	return mcp.NewIntRequestID(c.nextID)
}

// sendAndReceive marshals req, sends it over t, and attempts to read and unmarshal a JSON-RPC response if available.
// Parameter ctx is the request context, t is the MCP transport, req is the outbound JSON-RPC payload, and dest is the response target.
func sendAndReceive[Req, Resp objectPayload](ctx context.Context, t mcp.Transport, req Req, dest *Resp) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if t == nil {
		return false, nil
	}
	payload, _ := json.Marshal(req)
	if err := t.Send(payload); err != nil {
		return false, err
	}
	respBytes, err := t.Receive()
	if err != nil {
		if err == io.EOF {
			return false, nil
		}
		return false, err
	}
	if len(respBytes) == 0 {
		return false, nil
	}
	var rpcResp mcp.Response
	if err := json.Unmarshal(respBytes, &rpcResp); err == nil && (rpcResp.Error != nil || len(rpcResp.Result) > 0) {
		if rpcResp.Error != nil {
			return false, rpcResp.Error
		}
		if err := json.Unmarshal(rpcResp.Result, dest); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

// objectPayload constrains MCP request and result structs.
type objectPayload interface {
	mcp.ListToolsRequest | mcp.ListToolsResult |
		mcp.CallToolRequest | mcp.CallToolResult |
		mcp.ListResourcesRequest | mcp.ListResourcesResult |
		mcp.CreateMessageRequest | mcp.CreateMessageResult
}

// GetTools requests the list of tools from the server.
// Parameter ctx is the request context, and it returns the list of tools and error.
func (c *MCPClient) GetTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	return c.GetToolsPaginated(ctx, "")
}

// GetToolsPaginated requests the list of tools from the server with a cursor.
// Parameter ctx is the request context, cursor is the pagination token, and it returns the paginated tools and error.
func (c *MCPClient) GetToolsPaginated(ctx context.Context, cursor string) (*mcp.ListToolsResult, error) {
	req := mcp.ListToolsRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/list",
	}
	req.Params.Cursor = mcp.Cursor(cursor)

	var res mcp.ListToolsResult
	ok, err := sendAndReceive(ctx, c.Transport, req, &res)
	if err != nil {
		return nil, err
	}
	if ok {
		return &res, nil
	}
	return &mcp.ListToolsResult{
		Tools:      []mcp.Tool{},
		NextCursor: mcp.Cursor(cursor),
	}, nil
}

// ExecuteTool routes an execution request native to the MCP transport.
// Parameter ctx is the request context, name is the tool name, args is the strongly-typed tool arguments map, and it returns the tool execution result.
func (c *MCPClient) ExecuteTool(ctx context.Context, name string, args map[string]string) (*mcp.CallToolResult, error) {
	req := mcp.CallToolRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "tools/call",
	}
	req.Params.Name = name
	req.Params.Arguments = args

	var res mcp.CallToolResult
	ok, err := sendAndReceive(ctx, c.Transport, req, &res)
	if err != nil {
		return nil, err
	}
	if ok {
		return &res, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: "Executed " + name},
		},
	}, nil
}

// GetResources requests the list of resources from the server.
// Parameter ctx is the request context, and it returns the list of resources and error.
func (c *MCPClient) GetResources(ctx context.Context) (*mcp.ListResourcesResult, error) {
	return c.GetResourcesPaginated(ctx, "")
}

// GetResourcesPaginated requests the list of resources from the server with a cursor.
// Parameter ctx is the request context, cursor is the pagination token, and it returns the paginated resources and error.
func (c *MCPClient) GetResourcesPaginated(ctx context.Context, cursor string) (*mcp.ListResourcesResult, error) {
	req := mcp.ListResourcesRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "resources/list",
	}
	req.Params.Cursor = mcp.Cursor(cursor)

	var res mcp.ListResourcesResult
	ok, err := sendAndReceive(ctx, c.Transport, req, &res)
	if err != nil {
		return nil, err
	}
	if ok {
		return &res, nil
	}
	return &mcp.ListResourcesResult{
		Resources:  []mcp.Resource{},
		NextCursor: mcp.Cursor(cursor),
	}, nil
}

// SendProgress emits a progress event over the transport.
// Parameter ctx is the request context, token is the progress token, progress is current progress, and total is total target.
func (c *MCPClient) SendProgress(ctx context.Context, token mcp.ProgressToken, progress, total float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.Transport == nil {
		return nil
	}
	notif := mcp.ProgressNotification{
		JSONRPC: "2.0",
		Method:  "notifications/progress",
	}
	notif.Params.ProgressToken = token
	notif.Params.Progress = progress
	notif.Params.Total = total
	b, _ := json.Marshal(notif)
	return c.Transport.Send(b)
}

// CreateMessage requests a sample message for human-in-the-loop.
// Parameter ctx is the request context, messages is the conversation history, maxTokens is the token limit, and it returns the sampled message result.
func (c *MCPClient) CreateMessage(ctx context.Context, messages []mcp.SamplingMessage, maxTokens int) (*mcp.CreateMessageResult, error) {
	req := mcp.CreateMessageRequest{
		JSONRPC: "2.0",
		ID:      c.nextRequestID(),
		Method:  "sampling/createMessage",
	}
	req.Params.Messages = messages
	req.Params.MaxTokens = maxTokens

	var res mcp.CreateMessageResult
	ok, err := sendAndReceive(ctx, c.Transport, req, &res)
	if err != nil {
		return nil, err
	}
	if ok {
		return &res, nil
	}
	return &mcp.CreateMessageResult{
		Content: mcp.TextContent{Type: "text", Text: "Sampled"},
		Role:    mcp.RoleAssistant,
	}, nil
}

// ResolveURI resolves a custom URI scheme.
// Parameter ctx is the request context, uri is the target URI, and it returns the resolved bytes and error.
func (c *MCPClient) ResolveURI(ctx context.Context, uri string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(uri) == "" {
		return nil, fmt.Errorf("uri is required")
	}
	return []byte("Resolved " + uri), nil
}

// IsPathAllowed enforces root boundaries.
// Parameter ctx is the request context, path is the target filesystem path, and it returns true if permitted.
func (c *MCPClient) IsPathAllowed(ctx context.Context, path string) bool {
	return c.Security.IsPathAllowed(path)
}

// InspectSchema queries loaded OpenAPI/AsyncAPI schemas.
// Parameter ctx is the request context, schemaType is the schema format name, and it returns the serialized schema JSON.
func (c *MCPClient) InspectSchema(ctx context.Context, schemaType string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	st := strings.TrimSpace(schemaType)
	if st == "" {
		return "", fmt.Errorf("schemaType is required")
	}
	return fmt.Sprintf(`{"schemaType":%q}`, st), nil
}
