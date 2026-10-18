package cdd

import (
	"os"
	"testing"
)

func TestToDocsJSON(t *testing.T) {
	// Missing input
	if err := ToDocsJSON("", "", false, false); err == nil {
		t.Error("expected error for missing input")
	}

	// Invalid input
	if err := ToDocsJSON("missing.json", "", false, false); err == nil {
		t.Error("expected error for missing file")
	}

	// Create valid input
	inPath := "test_docs_input.json"
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}, "paths": {"/pets": {"get": {"operationId": "getPets"}}}}`), 0644)
	defer os.Remove(inPath)

	// Valid input to stdout
	if err := ToDocsJSON(inPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Valid input with wrapping options and to file
	outPath := "test_docs_output.json"
	if err := ToDocsJSON(inPath, outPath, true, true); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer os.Remove(outPath)

	// Invalid output path
	if err := ToDocsJSON(inPath, "/invalid/path/output.json", false, false); err == nil {
		t.Error("expected error for invalid output path")
	}

	// Invalid openapi spec (to trigger parse error)
	invalidIn := "test_docs_invalid.json"
	os.WriteFile(invalidIn, []byte(`{invalid`), 0644)
	defer os.Remove(invalidIn)
	if err := ToDocsJSON(invalidIn, "", false, false); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToDocsJSONCoverage(t *testing.T) {
	inPath := "test_docs_input2.json"
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}, "paths": {"/pets": {"get": {}}}}`), 0644)
	defer os.Remove(inPath)

	// test no operationId
	if err := ToDocsJSON(inPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test empty paths to trigger operations == nil
	emptyPath := "test_docs_empty.json"
	os.WriteFile(emptyPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}}`), 0644)
	defer os.Remove(emptyPath)
	if err := ToDocsJSON(emptyPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test Arazzo 1.1.0 spec input
	arazzoPath := "test_docs_arazzo.json"
	os.WriteFile(arazzoPath, []byte(`{"arazzo": "1.1.0", "info": {"title": "Wf", "version": "1.0"}, "sourceDescriptions": [{"name": "s", "url": "./o.json"}], "workflows": [{"workflowId": "loginWf", "steps": [{"stepId": "s1", "operationId": "op1"}]}]}`), 0644)
	defer os.Remove(arazzoPath)
	if err := ToDocsJSON(arazzoPath, "", false, false); err != nil {
		t.Errorf("unexpected error on Arazzo docs: %v", err)
	}
	if err := ToDocsJSON(arazzoPath, "", true, true); err != nil {
		t.Errorf("unexpected error on Arazzo docs noImports/noWrapping: %v", err)
	}

	// test directory passed as input file (triggers io.ReadAll error)
	if err := ToDocsJSON(t.TempDir(), "", false, false); err == nil {
		t.Error("expected error when passing directory to ToDocsJSON")
	}
}

func TestToDocsJSONEncodeErr(t *testing.T) {
	inPath := "test_docs_encode.json"
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "paths": {"/pets": {"get": {}}}}`), 0644)
	defer os.Remove(inPath)

	// Force json.Encode error by hijacking os.Stdout (outTarget defaults to stdout)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	w.Close() // closing the write end makes writing to it fail immediately

	err := ToDocsJSON(inPath, "", false, false)
	if err == nil {
		t.Error("Expected error writing to closed pipe")
	}

	os.Stdout = oldStdout
	r.Close()
}
