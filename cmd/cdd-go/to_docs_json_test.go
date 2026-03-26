package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunToDocsJSON(t *testing.T) {
	// Create a dummy openapi.json file for testing
	tempDir := t.TempDir()
	openAPIPath := filepath.Join(tempDir, "openapi.json")

	dummyOpenAPI := `{
  "openapi": "3.0.0",
  "info": { "title": "Test API", "version": "1.0.0" },
  "paths": {
    "/pets": {
      "get": {
        "operationId": "listPets"
      }
    }
  }
}`
	err := os.WriteFile(openAPIPath, []byte(dummyOpenAPI), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy OpenAPI file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		wantImports bool
		wantWrapper bool
	}{
		{
			name:        "Default behavior",
			args:        []string{"-i", openAPIPath},
			wantImports: true,
			wantWrapper: true,
		},
		{
			name:        "No imports",
			args:        []string{"-i", openAPIPath, "--no-imports"},
			wantImports: false,
			wantWrapper: true,
		},
		{
			name:        "No wrapping",
			args:        []string{"-i", openAPIPath, "--no-wrapping"},
			wantImports: true,
			wantWrapper: false,
		},
		{
			name:        "No imports and no wrapping",
			args:        []string{"-i", openAPIPath, "--no-imports", "--no-wrapping"},
			wantImports: false,
			wantWrapper: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Redirect stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := runToDocsJSON(tt.args)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Fatalf("runToDocsJSON failed: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result []DocsJSONOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("Failed to unmarshal output: %v\nOutput was: %s", err, output)
			}

			if len(result) != 1 {
				t.Fatalf("Expected 1 language output, got %d", len(result))
			}

			langResult := result[0]
			if langResult.Language != "go" {
				t.Errorf("Expected language 'go', got '%s'", langResult.Language)
			}

			if len(langResult.Operations) != 1 {
				t.Fatalf("Expected 1 operation, got %d", len(langResult.Operations))
			}

			op := langResult.Operations[0]
			if op.Method != "GET" {
				t.Errorf("Expected method 'GET', got '%s'", op.Method)
			}
			if op.Path != "/pets" {
				t.Errorf("Expected path '/pets', got '%s'", op.Path)
			}
			if op.OperationId != "listPets" {
				t.Errorf("Expected operationId 'listPets', got '%s'", op.OperationId)
			}

			code := op.Code
			if !strings.Contains(code.Snippet, "listPets") {
				t.Errorf("Expected snippet to contain 'listPets', got '%s'", code.Snippet)
			}

			if tt.wantImports {
				if code.Imports == nil {
					t.Errorf("Expected imports to be present")
				}
			} else {
				if code.Imports != nil {
					t.Errorf("Expected imports to be absent, got '%s'", *code.Imports)
				}
			}

			if tt.wantWrapper {
				if code.WrapperStart == nil || code.WrapperEnd == nil {
					t.Errorf("Expected wrappers to be present")
				}
			} else {
				if code.WrapperStart != nil || code.WrapperEnd != nil {
					t.Errorf("Expected wrappers to be absent")
				}
			}
		})
	}
}

func TestRunToDocsJSONErrors(t *testing.T) {
	err := runToDocsJSON([]string{"-invalid-flag"})
	if err == nil {
		t.Errorf("expected error for invalid flag")
	}

	err = runToDocsJSON([]string{})
	if err == nil {
		t.Errorf("expected error for missing input")
	}

	err = runToDocsJSON([]string{"-i", "missing-file.json"})
	if err == nil {
		t.Errorf("expected error for missing file")
	}

	dir := t.TempDir()
	invalidFile := filepath.Join(dir, "invalid.json")
	os.WriteFile(invalidFile, []byte("{invalid json"), 0644)
	err = runToDocsJSON([]string{"-i", invalidFile})
	if err == nil {
		t.Errorf("expected error for invalid json")
	}
}

func TestRunToDocsJSONEmpty(t *testing.T) {
	dir := t.TempDir()
	openAPIPath := filepath.Join(dir, "openapi.json")
	os.WriteFile(openAPIPath, []byte(`{
  "openapi": "3.0.0",
  "info": { "title": "Test API", "version": "1.0.0" },
  "paths": {}
}`), 0644)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	err := runToDocsJSON([]string{"-i", openAPIPath})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test encode error by passing a bad stdout? No, json.NewEncoder(os.Stdout).Encode(result)
	// We can just close os.Stdout BEFORE running it, then Encode will fail!

}

func TestRunToDocsJSONEncodeError(t *testing.T) {
	dir := t.TempDir()
	openAPIPath := filepath.Join(dir, "openapi.json")
	os.WriteFile(openAPIPath, []byte(`{
  "openapi": "3.0.0",
  "info": { "title": "Test API", "version": "1.0.0" },
  "paths": {}
}`), 0644)

	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	w.Close() // Close immediately so Encode fails!

	err := runToDocsJSON([]string{"-i", openAPIPath})

	os.Stdout = oldStdout

	if err == nil {
		t.Errorf("expected error from json Encode")
	}
}

func TestRunToDocsJSONNoOpID(t *testing.T) {
	dir := t.TempDir()
	openAPIPath := filepath.Join(dir, "openapi.json")
	os.WriteFile(openAPIPath, []byte(`{
  "openapi": "3.0.0",
  "info": { "title": "Test API", "version": "1.0.0" },
  "paths": {
    "/pets": {
      "get": {}
    }
  }
}`), 0644)

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runToDocsJSON([]string{"-i", openAPIPath})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result []DocsJSONOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to unmarshal output: %v", err)
	}

	if result[0].Operations[0].OperationId != "request" {
		t.Errorf("expected operationId \"request\", got \"%s\"", result[0].Operations[0].OperationId)
	}
}
