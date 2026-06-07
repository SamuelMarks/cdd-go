package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaToolsStructs(t *testing.T) {
	// Test Tool
	tool := Tool{Name: "my_tool", Description: "does a thing", InputSchema: map[string]interface{}{"type": "object"}}
	b, _ := json.Marshal(tool)
	if !strings.Contains(string(b), `"my_tool"`) {
		t.Errorf("expected tool name")
	}

	// Test CallToolRequest
	req := CallToolRequest{Method: "tools/call"}
	req.Params.Name = "my_tool"
	b, _ = json.Marshal(req)
	if !strings.Contains(string(b), `"my_tool"`) {
		t.Errorf("expected tool name in request")
	}

	// Test CallToolResult
	res := CallToolResult{IsError: true}
	res.Content = []Content{TextContent{Type: "text", Text: "error msg"}}
	b, _ = json.Marshal(res)
	if !strings.Contains(string(b), `"error msg"`) || !strings.Contains(string(b), `"isError":true`) {
		t.Errorf("expected error in result")
	}

	// Test ToolListChangedNotification
	not := ToolListChangedNotification{Method: "notifications/tools/list_changed"}
	b, _ = json.Marshal(not)
	if !strings.Contains(string(b), `"notifications/tools/list_changed"`) {
		t.Errorf("expected method")
	}

	// Test ListToolsRequest
	listReq := ListToolsRequest{Method: "tools/list"}
	b, _ = json.Marshal(listReq)
	if !strings.Contains(string(b), `"tools/list"`) {
		t.Errorf("expected method")
	}

	// Test ListToolsResult
	listRes := ListToolsResult{NextCursor: "c1"}
	listRes.Tools = append(listRes.Tools, tool)
	b, _ = json.Marshal(listRes)
	if !strings.Contains(string(b), `"c1"`) || !strings.Contains(string(b), `"my_tool"`) {
		t.Errorf("expected nextCursor and tool")
	}
}
