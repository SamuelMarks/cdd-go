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
	ToDocsJSON(inPath, "", false, false)

	// test empty paths to trigger operations == nil
	emptyPath := "test_docs_empty.json"
	os.WriteFile(emptyPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}}`), 0644)
	defer os.Remove(emptyPath)
	ToDocsJSON(emptyPath, "", false, false)
}

type badWriter struct{}

func (w badWriter) Write(p []byte) (n int, err error) {
	return 0, os.ErrClosed
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
