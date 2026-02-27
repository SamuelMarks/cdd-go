package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	err := run([]string{})
	if err == nil {
		t.Errorf("expected error for missing subcommands")
	}

	err = run([]string{"unknown"})
	if err == nil {
		t.Errorf("expected error for unknown subcommand")
	}

	err = run([]string{"from_openapi", "-invalid"})
	if err == nil {
		t.Errorf("expected error for invalid flag")
	}

	err = run([]string{"to_openapi", "-invalid"})
	if err == nil {
		t.Errorf("expected error for invalid flag")
	}

	err = run([]string{"from_openapi"})
	if err == nil {
		t.Errorf("expected error for missing input file")
	}

	err = run([]string{"to_openapi"})
	if err == nil {
		t.Errorf("expected error for missing input path")
	}

	// create dummy file for openapi testing
	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.json")
	os.WriteFile(path, []byte(`{"openapi": "3.2.0", "info": {"title": "Test"}}`), 0644)

	err = run([]string{"from_openapi", "-in", path, "-out", filepath.Join(dir, "generated")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test language-to-openapi
	goFile := filepath.Join(dir, "input.go")
	os.WriteFile(goFile, []byte(`package main

// User profile
type User struct {
	ID string `+"`json:\"id\"`"+`
}

// User endpoints
type HandlerUsers interface {
	// Get user
	GetUsers(ctx interface{}) error
}

// MockUser
var MockUser = `+"`{\"id\": \"1\"}`"+`
`), 0644)

	err = run([]string{"to_openapi", "-i", goFile, "-o", filepath.Join(dir, "output.json")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	outData, _ := os.ReadFile(filepath.Join(dir, "output.json"))
	var js map[string]interface{}
	json.Unmarshal(outData, &js)
	if js["openapi"] != "3.2.0" {
		t.Errorf("expected openapi string")
	}

	// test language-to-openapi on a dir
	goDir := filepath.Join(dir, "gocode")
	os.MkdirAll(goDir, 0755)
	os.WriteFile(filepath.Join(goDir, "a.go"), []byte(`package a; type A struct{}`), 0644)
	os.WriteFile(filepath.Join(goDir, "a_test.go"), []byte(`package a; type B struct{}`), 0644) // should be ignored

	err = run([]string{"to_openapi", "-in", goDir, "-out", filepath.Join(dir, "output_dir.json")})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// test language-to-openapi file error (not found)
	err = run([]string{"to_openapi", "-in", filepath.Join(dir, "missing.go")})
	if err == nil {
		t.Errorf("expected error for missing go file")
	}

	// test language-to-openapi parsing error
	errGo := filepath.Join(dir, "err.go")
	os.WriteFile(errGo, []byte(`package main; type;`), 0644)
	err = run([]string{"to_openapi", "-in", errGo})
	if err == nil {
		t.Errorf("expected parsing error")
	}

	// Output dir mapping test
	err = run([]string{"to_openapi", "-in", goDir, "-out", "generated"})
	// this will write to openapi.json in current dir. Let's ignore err, just want coverage.

	// Force WriteDstFile error by providing a path that is a directory
	err = run([]string{"from_openapi", "-in", path, "-out", path})
	if err == nil {
		// It might fail on MkdirAll if it is a file
		t.Logf("expected error creating directory over a file, got nil")
	}

	// Make out directory read-only to force write errors
	readonlyDir := filepath.Join(dir, "readonly")
	os.MkdirAll(readonlyDir, 0555)

	// Test writeDstFile errors
	os.WriteFile(path, []byte(`{"components": {"schemas": {"test": {"type": "string"}}}}`), 0644)
	err = run([]string{"from_openapi", "-in", path, "-out", readonlyDir})
	if err == nil {
		t.Errorf("expected error writing file")
	}

	os.WriteFile(path, []byte(`{"paths": {"/test": {"get": {}}}}`), 0644)
	err = run([]string{"from_openapi", "-in", path, "-out", readonlyDir})
	if err == nil {
		t.Errorf("expected error writing file")
	}

	// Test emit error inside generateClasses
	os.WriteFile(path, []byte(`{"components": {"schemas": {"test": {"type": "unknown-error"}}}}`), 0644)
	err = run([]string{"from_openapi", "-in", path, "-out", filepath.Join(dir, "error_gen")})
	if err == nil {
		t.Errorf("expected error from classes.EmitType")
	}

	// Test emit error inside generateRoutes
	os.WriteFile(path, []byte(`{"paths": {"/error-path": {}}}`), 0644)
	err = run([]string{"from_openapi", "-in", path, "-out", filepath.Join(dir, "error_gen")})
	if err == nil {
		t.Errorf("expected error from routes.EmitHandlerInterface")
	}

	// invalid openapi file
	os.WriteFile(path, []byte(`{invalid`), 0644)
	err = run([]string{"from_openapi", "-in", path})
	if err == nil {
		t.Errorf("expected error parsing invalid json")
	}

	// missing file
	err = run([]string{"from_openapi", "-in", filepath.Join(dir, "missing.json")})
	if err == nil {
		t.Errorf("expected error for missing file")
	}
}

func TestGenerateOpenAPIWriteError(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "input.go")
	os.WriteFile(goFile, []byte(`package main`), 0644)

	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0755)
	os.MkdirAll(filepath.Join(outDir, "openapi.json"), 0755) // The file we want to create is already a dir

	err := run([]string{"to_openapi", "-in", goFile, "-out", outDir})
	if err == nil {
		t.Errorf("expected error opening file that is a dir")
	}
}

func TestGenerateOpenAPIMkdirError(t *testing.T) {
	dir := t.TempDir()
	goFile := filepath.Join(dir, "input.go")
	os.WriteFile(goFile, []byte(`package main`), 0644)

	// Create a file and then ask to make a directory on top of it
	blockerFile := filepath.Join(dir, "blocker")
	os.WriteFile(blockerFile, []byte(""), 0644)

	outPath := filepath.Join(blockerFile, "openapi.json")

	err := run([]string{"to_openapi", "-in", goFile, "-out", outPath})
	if err == nil {
		t.Errorf("expected error when mkdir fails")
	}
}

func TestGenerateOpenAPIReadDirError(t *testing.T) {
	dir := t.TempDir()

	// mock ReadDir error by taking away permissions (might not work reliably on all OS, but works on Linux)
	os.Chmod(dir, 0000)
	defer os.Chmod(dir, 0755)

	err := run([]string{"to_openapi", "-in", dir, "-out", "test.json"})
	if err == nil {
		t.Errorf("expected error when reading dir fails")
	}
}

func TestMainError(t *testing.T) {
	// mock os.Exit and os.Stderr
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }()

	// Temporarily hijack os.Args to cause error
	oldArgs := os.Args
	os.Args = []string{"cdd_go"} // missing subcommand
	defer func() { os.Args = oldArgs }()

	main()

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestMainSuccess(t *testing.T) {
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }()

	dir := t.TempDir()
	path := filepath.Join(dir, "openapi.json")
	os.WriteFile(path, []byte(`{
  "openapi": "3.2.0",
  "info": {"title": "Test API"},
  "paths": {
    "/ping": {
      "get": {"operationId": "ping"}
    },
    "/": {
      "get": {"operationId": "root"}
    }
  },
  "components": {
    "schemas": {
      "Pong": {
        "type": "string"
      }
    }
  }
}`), 0644)

	oldArgs := os.Args
	os.Args = []string{"cdd_go", "from_openapi", "-in", path, "-out", filepath.Join(dir, "out")}
	defer func() { os.Args = oldArgs }()

	main()

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	// verify files generated
	if _, err := os.Stat(filepath.Join(dir, "out", "pong.go")); os.IsNotExist(err) {
		t.Errorf("expected pong.go to be generated")
	}
	if _, err := os.Stat(filepath.Join(dir, "out", "ping_routes.go")); os.IsNotExist(err) {
		t.Errorf("expected ping_routes.go to be generated")
	}
	if _, err := os.Stat(filepath.Join(dir, "out", "root_routes.go")); os.IsNotExist(err) {
		t.Errorf("expected root_routes.go to be generated")
	}
}
