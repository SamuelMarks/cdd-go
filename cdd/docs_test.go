package cdd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// failingDocWriter is a test io.Writer that always returns an error.
type failingDocWriter struct{}

func (failingDocWriter) Write(p []byte) (int, error) {
	return 0, errors.New("mock write failure")
}

func TestToDocsJSON(t *testing.T) {
	oldOut := docsDefaultStdout
	defer func() { docsDefaultStdout = oldOut }()
	var outBuf bytes.Buffer
	docsDefaultStdout = &outBuf

	// Missing input
	if err := ToDocsJSON("", "", false, false); err == nil {
		t.Error("expected error for missing input")
	}

	// Invalid input
	if err := ToDocsJSON("missing.json", "", false, false); err == nil {
		t.Error("expected error for missing file")
	}

	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "test_docs_input.json")
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}, "paths": {"/pets": {"get": {"operationId": "getPets"}}}}`), 0644)

	// Valid input to default writer
	if err := ToDocsJSON(inPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(outBuf.String(), "ClientPets") || !strings.Contains(outBuf.String(), "GetPets") {
		t.Errorf("expected ClientPets and GetPets in output, got %s", outBuf.String())
	}

	// Valid input with wrapping options and to file
	outPath := filepath.Join(tmpDir, "test_docs_output.json")
	if err := ToDocsJSON(inPath, outPath, true, true); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Valid input with out as an existing directory path
	outDir := t.TempDir()
	if err := ToDocsJSON(inPath, outDir, false, false); err != nil {
		t.Errorf("unexpected error when out is a directory: %v", err)
	}
	dirDocsBytes, err := os.ReadFile(filepath.Join(outDir, "docs.json"))
	if err != nil {
		t.Fatalf("expected docs.json inside outDir: %v", err)
	}
	if !strings.Contains(string(dirDocsBytes), "ClientPets") {
		t.Errorf("expected ClientPets in docs.json, got %s", string(dirDocsBytes))
	}

	// Invalid output path
	if err := ToDocsJSON(inPath, filepath.Join(tmpDir, "nonexistent_dir", "output.json"), false, false); err == nil {
		t.Error("expected error for invalid output path")
	}

	// Invalid openapi spec (to trigger parse error)
	invalidIn := filepath.Join(tmpDir, "test_docs_invalid.json")
	os.WriteFile(invalidIn, []byte(`{invalid`), 0644)
	if err := ToDocsJSON(invalidIn, "", false, false); err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestToDocsJSONCoverage(t *testing.T) {
	oldOut := docsDefaultStdout
	defer func() { docsDefaultStdout = oldOut }()
	var outBuf bytes.Buffer
	docsDefaultStdout = &outBuf

	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "test_docs_input2.json")
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}, "paths": {"/pets": {"get": {}}}}`), 0644)

	// test no operationId
	if err := ToDocsJSON(inPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test empty paths to trigger operations == nil
	emptyPath := filepath.Join(tmpDir, "test_docs_empty.json")
	os.WriteFile(emptyPath, []byte(`{"openapi": "3.0.0", "info": {"title": "Test", "version": "1.0"}}`), 0644)
	if err := ToDocsJSON(emptyPath, "", false, false); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test OpenAPI 3.2.0 webhooks support
	webhookPath := filepath.Join(tmpDir, "test_docs_webhooks.json")
	os.WriteFile(webhookPath, []byte(`{"openapi": "3.2.0", "info": {"title": "Test", "version": "1.0"}, "webhooks": {"newPet": {"post": {"operationId": "onNewPet"}}, "deletedPet": {"delete": {"operationId": "onDeletedPet"}}}}`), 0644)
	var whBuf bytes.Buffer
	if err := ToDocsJSONWriter(webhookPath, &whBuf, false, false); err != nil {
		t.Errorf("unexpected error on webhooks docs: %v", err)
	}
	if !strings.Contains(whBuf.String(), "OnNewPet") || !strings.Contains(whBuf.String(), "OnDeletedPet") {
		t.Errorf("expected webhook operations in output, got %s", whBuf.String())
	}

	// test Arazzo 1.1.0 spec input
	arazzoPath := filepath.Join(tmpDir, "test_docs_arazzo.json")
	os.WriteFile(arazzoPath, []byte(`{"arazzo": "1.1.0", "info": {"title": "Wf", "version": "1.0"}, "sourceDescriptions": [{"name": "s", "url": "./o.json"}], "workflows": [{"workflowId": "loginWf", "steps": [{"stepId": "s1", "operationId": "op1"}]}]}`), 0644)
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
	oldOut := docsDefaultStdout
	defer func() { docsDefaultStdout = oldOut }()
	docsDefaultStdout = failingDocWriter{}

	tmpDir := t.TempDir()
	inPath := filepath.Join(tmpDir, "test_docs_encode.json")
	os.WriteFile(inPath, []byte(`{"openapi": "3.0.0", "paths": {"/pets": {"get": {}}}}`), 0644)

	if err := ToDocsJSON(inPath, "", false, false); err == nil {
		t.Error("Expected error writing to failingDocWriter")
	}
}
