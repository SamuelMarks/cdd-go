package cdd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunServeMCPStdio(t *testing.T) {
	// Mock os.Stdin and os.Stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdin = r
	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	// Write inputs
	inputs := []string{
		`{"jsonrpc": "2.0", "method": "initialize", "id": 1}`,
		`{"jsonrpc": "2.0", "method": "tools/list", "id": 2}`,
		`{"jsonrpc": "2.0", "method": "unknown", "id": 3}`,
	}
	for _, input := range inputs {
		w.Write([]byte(input + "\n"))
	}
	w.Close()

	// Run MCP in background
	done := make(chan struct{})
	go func() {
		ServeMCPStdio([]string{})
		outW.Close()
		close(done)
	}()

	var buf bytes.Buffer
	io.Copy(&buf, outR)
	<-done

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	output := buf.String()
	if !strings.Contains(output, `"protocolVersion":"1.0.0"`) {
		t.Errorf("Expected initialize response")
	}
	if !strings.Contains(output, `"cdd_generate_sdk"`) {
		t.Errorf("Expected tools/list response")
	}
	if !strings.Contains(output, `"code":-32601`) {
		t.Errorf("Expected method not found response")
	}
}

func TestRunServeMCPStdioInvalidJSON(t *testing.T) {
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	w.Write([]byte(`{invalid_json` + "\n"))
	w.Close()

	err := ServeMCPStdio([]string{})
	if err != nil {
		t.Errorf("Did not expect error on invalid json, should just continue to EOF")
	}

	os.Stdin = oldStdin
}

func TestRunServeMCPStdioToolsCall(t *testing.T) {
	// Mock os.Stdin and os.Stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdin = r
	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	// Write inputs
	inputs := []string{
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 1, "params": {"name": "cdd_generate_sdk", "arguments": {"input": "../test_spec.json", "output": "../test_out/my_mcp_sdk"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 2, "params": {"name": "cdd_sync_schema", "arguments": {"input": "../test_out/my_mcp_sdk", "output": "../test_out/my_mcp_schema.json"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 3, "params": {"name": "unknown_tool", "arguments": {}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 4, "params": {"name": "cdd_generate_sdk", "arguments": {"input": "does-not-exist.json", "output": "../test_out/missing"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 5, "params": {"name": "cdd_sync_schema", "arguments": {"input": "", "output": ""}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 6, "params": []}`,
	}
	for _, input := range inputs {
		w.Write([]byte(input + "\n"))
	}
	w.Close()

	done := make(chan struct{})
	go func() {
		ServeMCPStdio([]string{})
		outW.Close()
		close(done)
	}()

	var buf bytes.Buffer
	io.Copy(&buf, outR)
	<-done

	os.Stdin = oldStdin
	os.Stdout = oldStdout

	output := buf.String()
	if !strings.Contains(output, `"SDK successfully generated"`) {
		t.Errorf("Expected SDK generated response, got %s", output)
	}
	if !strings.Contains(output, `"OpenAPI schema successfully extracted"`) {
		t.Errorf("Expected schema extracted response, got %s", output)
	}
	if !strings.Contains(output, `"Method not found"`) {
		t.Errorf("Expected method not found response, got %s", output)
	}
	if !strings.Contains(output, `"code":-32603`) {
		t.Errorf("Expected error code -32603 response, got %s", output)
	}
	if !strings.Contains(output, `"Invalid params"`) {
		t.Errorf("Expected Invalid params response, got %s", output)
	}
}
