package main

import (
	"bytes"
	"encoding/json"
	"github.com/SamuelMarks/cdd-go/cdd"
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
		t.Fatalf("failed to write dummy openapi: %v", err)
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
				t.Fatalf("unexpected error: %v", err)
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result []cdd.DocsJSONOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("failed to unmarshal output: %v", err)
			}

			if len(result) != 1 {
				t.Fatalf("expected 1 language result, got %d", len(result))
			}

			langResult := result[0]
			if langResult.Language != "go" {
				t.Errorf("expected language 'go', got %q", langResult.Language)
			}

			if len(langResult.Operations) != 1 {
				t.Fatalf("expected 1 operation, got %d", len(langResult.Operations))
			}

			op := langResult.Operations[0]
			if op.Method != "GET" {
				t.Errorf("expected method GET, got %q", op.Method)
			}
			if op.Path != "/pets" {
				t.Errorf("expected path /pets, got %q", op.Path)
			}
			if op.OperationId != "listPets" {
				t.Errorf("expected operationId listPets, got %q", op.OperationId)
			}

			code := op.Code
			if !strings.Contains(code.Snippet, "listPets") {
				t.Errorf("expected snippet to contain listPets, got %q", code.Snippet)
			}

			if tt.wantImports {
				if code.Imports == nil {
					t.Errorf("expected non-nil Imports")
				}
			} else {
				if code.Imports != nil {
					t.Errorf("expected nil Imports, got %q", *code.Imports)
				}
			}

			if tt.wantWrapper {
				if code.WrapperStart == nil || code.WrapperEnd == nil {
					t.Errorf("expected non-nil WrapperStart and WrapperEnd")
				}
			} else {
				if code.WrapperStart != nil || code.WrapperEnd != nil {
					t.Errorf("expected nil WrapperStart and WrapperEnd")
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
		t.Errorf("expected error for missing input flag")
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
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runToDocsJSON([]string{"-i", openAPIPath})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if !strings.Contains(buf.String(), `"operations": []`) {
		t.Errorf("expected empty operations array, got %s", buf.String())
	}
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
		t.Errorf("expected error writing to closed stdout")
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
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result []cdd.DocsJSONOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	if result[0].Operations[0].OperationId != "request" {
		t.Errorf("expected default operationId 'request', got %q", result[0].Operations[0].OperationId)
	}
}
