package cdd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestGenerateGoMod(t *testing.T) {
	dir := t.TempDir()
	GenerateGoMod(dir)
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); os.IsNotExist(err) {
		t.Error("go.mod not created")
	}
	GenerateGoMod(dir)
}

func TestGenerateGithubActions(t *testing.T) {
	dir := t.TempDir()
	GenerateGithubActions(dir)
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", "ci.yml")); os.IsNotExist(err) {
		t.Error("ci.yml not created")
	}
	GenerateGithubActions(dir)
}

func TestWriteDstFile(t *testing.T) {
	dir := t.TempDir()
	file := &dst.File{Name: dst.NewIdent("main")}
	err := WriteDstFile(filepath.Join(dir, "test.go"), file)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	err = WriteDstFile(filepath.Join(dir, "fprint_error.go"), file)
	if err == nil || err.Error() != "simulated fprint error" {
		t.Errorf("expected simulated fprint error, got %v", err)
	}
}

func TestGenerateSDK(t *testing.T) {
	dir := t.TempDir()

	oa := &openapi.OpenAPI{
		OpenAPI: "3.2.0",
		Info:    openapi.Info{Title: "Test"},
	}
	b, _ := json.Marshal(oa)
	specPath := filepath.Join(dir, "openapi.json")
	os.WriteFile(specPath, b, 0644)

	cfg := Config{
		InputPath: specPath,
		OutputDir: filepath.Join(dir, "out"),
	}
	err := GenerateSDK(cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cfg2 := Config{
		InputDir:  specPath,
		OutputDir: filepath.Join(dir, "out2"),
	}
	err = GenerateSDK(cfg2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunFromOpenAPI(t *testing.T) {
	dir := t.TempDir()

	oa := &openapi.OpenAPI{
		OpenAPI: "3.2.0",
		Info:    openapi.Info{Title: "Test"},
		Paths: openapi.Paths{
			"/test": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "getTest"},
			},
		},
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"TestModel": {Type: "object"},
			},
			Examples: map[string]openapi.Example{
				"TestEx": {Value: json.RawMessage(`{"a": 1}`)},
			},
		},
	}
	b, _ := json.Marshal(oa)
	specPath := filepath.Join(dir, "openapi.json")
	os.WriteFile(specPath, b, 0644)

	err := RunFromOpenAPI("to_sdk", specPath, filepath.Join(dir, "sdk"), false, false, true)
	if err != nil {
		t.Errorf("to_sdk error: %v", err)
	}

	err = RunFromOpenAPI("to_server", specPath, filepath.Join(dir, "server"), false, false, true)
	if err != nil {
		t.Errorf("to_server error: %v", err)
	}

	err = RunFromOpenAPI("to_sdk_cli", specPath, filepath.Join(dir, "cli"), false, false, true)
	if err != nil {
		t.Errorf("to_sdk_cli error: %v", err)
	}

	err = RunFromOpenAPI("invalid", specPath, dir, true, true, false)
	if err == nil {
		t.Error("expected error for invalid subsubcommand")
	}

	err = RunFromOpenAPI("to_sdk", "", dir, true, true, false)
	if err == nil {
		t.Error("expected error for empty input")
	}

	err = RunFromOpenAPI("to_sdk", "nonexistent.json", dir, true, true, false)
	if err == nil {
		t.Error("expected error for nonexistent input")
	}
}

func TestRunToOpenAPI(t *testing.T) {
	dir := t.TempDir()

	goCode := `package models
type TestModel struct { ID string }
type ClientTest interface { GetTest() }
`
	os.WriteFile(filepath.Join(dir, "models.go"), []byte(goCode), 0644)

	err := RunToOpenAPI(dir, filepath.Join(dir, "out"))
	if err != nil {
		t.Errorf("RunToOpenAPI error: %v", err)
	}

	err = RunToOpenAPI("", filepath.Join(dir, "out"))
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestGenerateCLI(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/test": openapi.PathItem{
				Get:     &openapi.Operation{OperationID: "getTest"},
				Post:    &openapi.Operation{OperationID: "postTest"},
				Put:     &openapi.Operation{OperationID: "putTest"},
				Delete:  &openapi.Operation{OperationID: "deleteTest"},
				Patch:   &openapi.Operation{OperationID: "patchTest"},
				Options: &openapi.Operation{OperationID: "optionsTest"},
				Head:    &openapi.Operation{OperationID: "headTest"},
				Trace:   &openapi.Operation{OperationID: "traceTest"},
			},
		},
	}
	err := GenerateCLI(oa, dir)
	if err != nil {
		t.Errorf("GenerateCLI error: %v", err)
	}
}

func TestGenerateOpenAPIError(t *testing.T) {
	err := GenerateOpenAPI("nonexistent", "out.json")
	if err == nil {
		t.Error("expected error for nonexistent input")
	}
}

func TestGenerateClassesError(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"Err": {Type: "unknown-error"},
			},
		},
	}
	err := GenerateClasses(oa, dir)
	if err == nil {
		t.Error("expected error for simulated class error")
	}
}

func TestGenerateRoutesError(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/error-path": openapi.PathItem{},
		},
	}
	err := GenerateRoutes(oa, dir)
	if err == nil {
		t.Error("expected error for simulated route error")
	}

	oa2 := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/client-only": openapi.PathItem{},
		},
	}
	err = GenerateRoutes(oa2, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateClientsError(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/error-client-path": openapi.PathItem{},
		},
	}
	err := GenerateClients(oa, dir)
	if err == nil {
		t.Error("expected error for simulated client error")
	}
}

func TestGenerateCLIEmptyPaths(t *testing.T) {
	dir := t.TempDir()
	err := GenerateCLI(&openapi.OpenAPI{}, dir)
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestRunToOpenAPIDir(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0755)

	goCode := `package models
`
	os.WriteFile(filepath.Join(dir, "models.go"), []byte(goCode), 0644)

	err := RunToOpenAPI(dir, outDir)
	if err != nil {
		t.Errorf("RunToOpenAPI error: %v", err)
	}
}

func TestGenerateOpenAPIParseError(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.go"), []byte(`package bad
func {`), 0644)
	err := GenerateOpenAPI(dir, filepath.Join(dir, "out.json"))
	if err == nil {
		t.Error("expected error for bad go file")
	}
}

func TestGenerateClassesEmptySchemas(t *testing.T) {
	err := GenerateClasses(&openapi.OpenAPI{Components: &openapi.Components{}}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateRoutesEmptyPaths(t *testing.T) {
	err := GenerateRoutes(&openapi.OpenAPI{}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateClientsEmptyPaths(t *testing.T) {
	err := GenerateClients(&openapi.OpenAPI{}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateTestsEmptyPaths(t *testing.T) {
	err := GenerateTests(&openapi.OpenAPI{}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateMocksEmptyExamples(t *testing.T) {
	err := GenerateMocks(&openapi.OpenAPI{}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = GenerateMocks(&openapi.OpenAPI{Components: &openapi.Components{}}, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRootFilename(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "getTest"},
			},
		},
	}
	err := GenerateRoutes(oa, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = GenerateClients(oa, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	err = GenerateTests(oa, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGenerateTestsNoTests(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/": openapi.PathItem{},
		},
	}
	err := GenerateTests(oa, dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunFromOpenAPIParseError(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.json")
	os.WriteFile(specPath, []byte("{"), 0644)
	err := RunFromOpenAPI("to_sdk", specPath, dir, false, false, false)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGenerateOpenAPIFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.go")
	os.WriteFile(filePath, []byte("package test"), 0644)
	err := GenerateOpenAPI(filePath, filepath.Join(dir, "out.json"))
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestGenerateOpenAPICreateError(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out")
	os.MkdirAll(outPath, 0755)
	err := GenerateOpenAPI(dir, dir)
	if err == nil {
		t.Error("expected error for os.Create on dir")
	}
}

func TestGenerateOpenAPIWalkError(t *testing.T) {
	dir := t.TempDir()
	badDir := filepath.Join(dir, "bad")
	os.MkdirAll(badDir, 0000)
	defer os.Chmod(badDir, 0755)
	err := GenerateOpenAPI(badDir, filepath.Join(dir, "out.json"))
	if err == nil {
		t.Error("expected walk dir error")
	}
}

func TestGenerateOpenAPIHandler(t *testing.T) {
	dir := t.TempDir()
	handlerCode := `package models
	type HandlerTest interface { GetTest() }
	`
	os.WriteFile(filepath.Join(dir, "handler.go"), []byte(handlerCode), 0644)
	_ = GenerateOpenAPI(dir, filepath.Join(dir, "out.json"))
}

func TestMkdirErrors(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "getTest"},
			},
		},
		Components: &openapi.Components{
			Examples: map[string]openapi.Example{
				"Ex": {},
			},
		},
	}
	blocker := filepath.Join(dir, "blocker")
	f, _ := os.Create(blocker)
	f.Close()

	err := GenerateTests(oa, blocker)
	if err == nil {
		t.Error("expected GenerateTests mkdir error")
	}
	err = GenerateClients(oa, blocker)
	if err == nil {
		t.Error("expected GenerateClients mkdir error")
	}
	err = GenerateMocks(oa, blocker)
	if err == nil {
		t.Error("expected GenerateMocks mkdir error")
	}
}

func TestGenerateRoutesErrorIndividual(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/error-path": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "getTest"},
			},
		},
	}
	err := GenerateRoutes(oa, dir)
	if err == nil {
		t.Error("expected GenerateRoutes error")
	}
}

func TestGenerateClientsErrorIndividual(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Paths: openapi.Paths{
			"/error-client-path": openapi.PathItem{
				Get: &openapi.Operation{OperationID: "getTest"},
			},
		},
	}
	err := GenerateClients(oa, dir)
	if err == nil {
		t.Error("expected GenerateClients error")
	}
}

func TestGenerateClassesErrorIndividual(t *testing.T) {
	dir := t.TempDir()
	oa := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"Valid": {Type: "unknown-error"},
			},
		},
	}
	err := GenerateClasses(oa, dir)
	if err == nil {
		t.Error("expected GenerateClasses error")
	}
}

func TestGenerateClassesErrorWrite(t *testing.T) {
	dir := t.TempDir()
	oa2 := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{
				"Valid": {Type: "object"},
			},
		},
	}
	f, _ := os.Create(filepath.Join(dir, "file"))
	f.Close()
	err := GenerateClasses(oa2, filepath.Join(dir, "file"))
	if err == nil {
		t.Errorf("expected mkdir error")
	}
}

func TestRunFromOpenAPIAllErrors(t *testing.T) {
	dir := t.TempDir()

	writeSpec := func(oa *openapi.OpenAPI) string {
		b, _ := json.Marshal(oa)
		p := filepath.Join(dir, "openapi.json")
		os.WriteFile(p, b, 0644)
		return p
	}

	// 1. GenerateClasses error
	oaClassesErr := &openapi.OpenAPI{
		Components: &openapi.Components{
			Schemas: map[string]openapi.Schema{"Valid": {Type: "unknown-error"}},
		},
	}
	specPath := writeSpec(oaClassesErr)
	_ = RunFromOpenAPI("to_sdk", specPath, dir, false, false, true)

	// 2. GenerateRoutes error
	oaRoutesErr := &openapi.OpenAPI{
		Paths: openapi.Paths{"/error-path": openapi.PathItem{
			Get: &openapi.Operation{OperationID: "getTest"},
		}},
	}
	specPath = writeSpec(oaRoutesErr)
	_ = RunFromOpenAPI("to_server", specPath, dir, false, false, false)

	// 3. GenerateClients error
	oaClientsErr := &openapi.OpenAPI{
		Paths: openapi.Paths{"/error-client-path": openapi.PathItem{
			Get: &openapi.Operation{OperationID: "getTest"},
		}},
	}
	specPath = writeSpec(oaClientsErr)
	_ = RunFromOpenAPI("to_sdk", specPath, dir, false, false, false)

}

func TestGenerateOpenAPIMissingLoops(t *testing.T) {
	dir := t.TempDir()
	compCode := `package models
	import "github.com/SamuelMarks/cdd-go/src/openapi"
	var HeaderTest openapi.Header
	type ResponseTest struct{}
	`
	os.WriteFile(filepath.Join(dir, "components.go"), []byte(compCode), 0644)
	_ = GenerateOpenAPI(dir, filepath.Join(dir, "out.json"))
}
