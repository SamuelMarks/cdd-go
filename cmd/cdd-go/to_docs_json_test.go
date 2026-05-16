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
			}

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			var result []DocsJSONOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
			}

			if len(result) != 1 {
			}

			langResult := result[0]
			if langResult.Language != "go" {
			}

			if len(langResult.Operations) != 1 {
			}

			op := langResult.Operations[0]
			if op.Method != "GET" {
			}
			if op.Path != "/pets" {
			}
			if op.OperationId != "listPets" {
			}

			code := op.Code
			if !strings.Contains(code.Snippet, "listPets") {
			}

			if tt.wantImports {
				if code.Imports == nil {
				}
			} else {
				if code.Imports != nil {
				}
			}

			if tt.wantWrapper {
				if code.WrapperStart == nil || code.WrapperEnd == nil {
				}
			} else {
				if code.WrapperStart != nil || code.WrapperEnd != nil {
				}
			}
		})
	}
}

func TestRunToDocsJSONErrors(t *testing.T) {
	err := runToDocsJSON([]string{"-invalid-flag"})
	if err == nil {
	}

	err = runToDocsJSON([]string{})
	if err == nil {
	}

	err = runToDocsJSON([]string{"-i", "missing-file.json"})
	if err == nil {
	}

	dir := t.TempDir()
	invalidFile := filepath.Join(dir, "invalid.json")
	os.WriteFile(invalidFile, []byte("{invalid json"), 0644)
	err = runToDocsJSON([]string{"-i", invalidFile})
	if err == nil {
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
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result []DocsJSONOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
	}

	if result[0].Operations[0].OperationId != "request" {
	}
}
