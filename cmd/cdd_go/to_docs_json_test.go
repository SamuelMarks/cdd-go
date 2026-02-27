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
