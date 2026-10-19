package cdd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeMCPProtocol(t *testing.T) {
	nl := string(rune(10))
	inputs := strings.Join([]string{
		`{"jsonrpc": "2.0", "method": "initialize", "id": 1}`,
		`{"jsonrpc": "2.0", "method": "tools/list", "id": 2}`,
		`{"jsonrpc": "2.0", "method": "unknown", "id": 3}`,
	}, nl) + nl

	var outBuf, errBuf bytes.Buffer
	if err := ServeMCP(strings.NewReader(inputs), &outBuf, &errBuf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, `"protocolVersion":"1.0.0"`) {
		t.Errorf("Expected initialize response, got %s", output)
	}
	if !strings.Contains(output, `"cdd_generate_sdk"`) {
		t.Errorf("Expected tools/list response, got %s", output)
	}
	if !strings.Contains(output, `"code":-32601`) {
		t.Errorf("Expected method not found response, got %s", output)
	}
}

func TestRunServeMCPStdio(t *testing.T) {
	oldIn, oldOut, oldErr := mcpStdin, mcpStdout, mcpStderr
	defer func() {
		mcpStdin, mcpStdout, mcpStderr = oldIn, oldOut, oldErr
	}()

	mcpStdin = strings.NewReader(`{"jsonrpc": "2.0", "method": "initialize", "id": 1}` + string(rune(10)))
	var outBuf, errBuf bytes.Buffer
	mcpStdout = &outBuf
	mcpStderr = &errBuf

	if err := ServeMCPStdio(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(outBuf.String(), `"protocolVersion":"1.0.0"`) {
		t.Errorf("Expected initialize response from ServeMCPStdio")
	}
}

func TestServeMCPInvalidJSON(t *testing.T) {
	var outBuf, errBuf bytes.Buffer
	err := ServeMCP(strings.NewReader(`{invalid_json`+string(rune(10))), &outBuf, &errBuf)
	if err != nil {
		t.Errorf("Did not expect error on invalid json, got %v", err)
	}
	if !strings.Contains(errBuf.String(), "Error decoding request:") {
		t.Errorf("Expected diagnostic output in errBuf, got %q", errBuf.String())
	}
}

func TestServeMCPToolsCall(t *testing.T) {
	tmpDir := t.TempDir()
	oaSpec := filepath.Join(tmpDir, "spec.json")
	os.WriteFile(oaSpec, []byte(`{"openapi":"3.2.0","info":{"title":"T","version":"1.0"},"paths":{"/users":{"get":{"operationId":"getUsers","responses":{"200":{"description":"ok"}}}}}}`), 0644)

	arazzoSpec := filepath.Join(tmpDir, "arazzo.json")
	os.WriteFile(arazzoSpec, []byte(`{"arazzo":"1.1.0","info":{"title":"T","version":"1.0"},"sourceDescriptions":[{"name":"s","url":"./o.json"}],"workflows":[{"workflowId":"w","steps":[{"stepId":"s","operationId":"op"}]}]}`), 0644)

	sdkOut := filepath.Join(tmpDir, "my_mcp_sdk")
	schemaOut := filepath.Join(tmpDir, "my_mcp_schema.json")
	azSdkOut := filepath.Join(tmpDir, "az_sdk")
	azSpecOut := filepath.Join(tmpDir, "az_out.json")

	inputs := strings.Join([]string{
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 1, "params": {"name": "cdd_generate_sdk", "arguments": {"input": "` + oaSpec + `", "output": "` + sdkOut + `"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 2, "params": {"name": "cdd_sync_schema", "arguments": {"input": "` + sdkOut + `", "output": "` + schemaOut + `"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 3, "params": {"name": "unknown_tool", "arguments": {}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 4, "params": {"name": "cdd_generate_sdk", "arguments": {"input": "does-not-exist.json", "output": "` + filepath.Join(tmpDir, "missing") + `"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 5, "params": {"name": "cdd_sync_schema", "arguments": {"input": "", "output": ""}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 6, "params": []}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 7, "params": {"name": "cdd_generate_arazzo_sdk", "arguments": {"input": "` + arazzoSpec + `", "output": "` + azSdkOut + `"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 8, "params": {"name": "cdd_generate_arazzo_sdk", "arguments": {"input": "", "output": ""}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 9, "params": {"name": "cdd_sync_arazzo", "arguments": {"input": "` + azSdkOut + `", "output": "` + azSpecOut + `"}}}`,
		`{"jsonrpc": "2.0", "method": "tools/call", "id": 10, "params": {"name": "cdd_sync_arazzo", "arguments": {"input": "", "output": ""}}}`,
	}, string(rune(10))) + string(rune(10))

	var outBuf, errBuf bytes.Buffer
	if err := ServeMCP(strings.NewReader(inputs), &outBuf, &errBuf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := outBuf.String()
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
