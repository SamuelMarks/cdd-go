package mcp

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaToolsStructsAndRoundTrip(t *testing.T) {
	tool := Tool{
		Name:        "my_tool",
		Description: "does a thing",
		InputSchema: ToolInputSchema{
			Type: "object",
			Properties: map[string]ToolPropertySchema{
				"arg": {Type: "string", Description: "input arg"},
			},
			Required: []string{"arg"},
		},
	}
	b, err := json.Marshal(tool)
	if err != nil || !strings.Contains(string(b), `"my_tool"`) {
		t.Fatalf("expected tool name: %v", err)
	}
	var toolOut Tool
	if err := json.Unmarshal(b, &toolOut); err != nil || !reflect.DeepEqual(tool, toolOut) {
		t.Errorf("Tool round-trip mismatch: %+v vs %+v (%v)", tool, toolOut, err)
	}

	req := CallToolRequest{JSONRPC: "2.0", ID: NewIntRequestID(1), Method: "tools/call"}
	req.Params.Name = "my_tool"
	req.Params.Arguments = map[string]string{"arg": "val"}
	b, err = json.Marshal(req)
	if err != nil || !strings.Contains(string(b), `"tools/call"`) || !strings.Contains(string(b), `"val"`) {
		t.Fatalf("expected call tool request fields: %v", err)
	}
	var reqOut CallToolRequest
	if err := json.Unmarshal(b, &reqOut); err != nil || reqOut.Params.Name != "my_tool" || reqOut.Params.Arguments["arg"] != "val" {
		t.Errorf("CallToolRequest round-trip failed: %+v, %v", reqOut, err)
	}

	res := CallToolResult{IsError: false}
	res.Content = append(res.Content, TextContent{Type: "text", Text: "output"})
	b, err = json.Marshal(res)
	if err != nil || !strings.Contains(string(b), `"output"`) {
		t.Fatalf("expected call tool result output: %v", err)
	}
	var resOut CallToolResult
	if err := json.Unmarshal(b, &resOut); err != nil || len(resOut.Content) != 1 || resOut.Content[0].(TextContent).Text != "output" {
		t.Errorf("CallToolResult round-trip failed: %+v, %v", resOut, err)
	}
	if err := json.Unmarshal([]byte(`123`), &resOut); err == nil {
		t.Errorf("expected error for non-object CallToolResult json")
	}
	if err := json.Unmarshal([]byte(`{"content":[{"type":"unknown"}]}`), &resOut); err == nil {
		t.Errorf("expected error for invalid content item in CallToolResult")
	}

	not := ToolListChangedNotification{JSONRPC: "2.0", Method: "notifications/tools/list_changed"}
	b, err = json.Marshal(not)
	if err != nil || !strings.Contains(string(b), `"notifications/tools/list_changed"`) {
		t.Fatalf("expected notification method: %v", err)
	}
	var notOut ToolListChangedNotification
	if err := json.Unmarshal(b, &notOut); err != nil || notOut.Method != not.Method {
		t.Errorf("ToolListChangedNotification round-trip failed: %+v, %v", notOut, err)
	}

	listReq := ListToolsRequest{JSONRPC: "2.0", ID: NewIntRequestID(2), Method: "tools/list"}
	listReq.Params.Cursor = "c1"
	b, err = json.Marshal(listReq)
	if err != nil || !strings.Contains(string(b), `"tools/list"`) {
		t.Fatalf("expected list tools request: %v", err)
	}
	var listReqOut ListToolsRequest
	if err := json.Unmarshal(b, &listReqOut); err != nil || listReqOut.Params.Cursor != "c1" {
		t.Errorf("ListToolsRequest round-trip failed: %+v, %v", listReqOut, err)
	}

	listRes := ListToolsResult{NextCursor: "c1", Tools: []Tool{tool}}
	b, err = json.Marshal(listRes)
	if err != nil || !strings.Contains(string(b), `"c1"`) || !strings.Contains(string(b), `"my_tool"`) {
		t.Fatalf("expected nextCursor and tool: %v", err)
	}
	var listResOut ListToolsResult
	if err := json.Unmarshal(b, &listResOut); err != nil || !reflect.DeepEqual(listRes, listResOut) {
		t.Errorf("ListToolsResult round-trip mismatch: %+v vs %+v (%v)", listRes, listResOut, err)
	}
}
