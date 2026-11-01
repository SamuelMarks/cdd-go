package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestCLICommandsEndToEnd(t *testing.T) {
	dir, err := os.MkdirTemp("", "cdd-go-test-cli-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	specPath := filepath.Join(dir, "openapi.json")
	outDir := filepath.Join(dir, "generated")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatalf("failed to create outDir: %v", err)
	}

	specContent := `{
		"openapi": "3.2.0",
		"paths": {
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"summary": "Get User",
					"description": "Gets a user by ID"
				},
				"post": {
					"operationId": "createUser"
				},
				"put": {}
			}
		}
	}`
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	err = run([]string{"from_openapi", "to_sdk_cli", "-i", specPath, "-o", outDir})
	if err != nil {
		t.Fatalf("from_openapi to_sdk_cli failed: %v", err)
	}

	cliCode, err := os.ReadFile(filepath.Join(outDir, "sdk_cli.go"))
	if err != nil {
		t.Fatalf("failed to read sdk_cli.go: %v", err)
	}
	sCliCode := string(cliCode)

	if !strings.Contains(sCliCode, `var GetUserCmd = &cobra.Command`) {
		t.Errorf("expected GetUserCmd in sdk_cli.go, got:\n%s", sCliCode)
	}
	if !strings.Contains(sCliCode, `var CreateUserCmd = &cobra.Command`) {
		t.Errorf("expected CreateUserCmd in sdk_cli.go, got:\n%s", sCliCode)
	}
	if !strings.Contains(sCliCode, `var PutUsersIdCmd = &cobra.Command`) {
		t.Errorf("expected PutUsersIdCmd in sdk_cli.go, got:\n%s", sCliCode)
	}

	outSpec := filepath.Join(dir, "openapi_regen.json")
	err = run([]string{"to_openapi", "-i", outDir, "-o", outSpec})
	if err != nil {
		t.Fatalf("to_openapi failed: %v", err)
	}

	regenBytes, err := os.ReadFile(outSpec)
	if err != nil {
		t.Fatalf("failed to read regenerated spec: %v", err)
	}
	var oa openapi.OpenAPI
	if err := json.Unmarshal(regenBytes, &oa); err != nil {
		t.Fatalf("failed to unmarshal regenerated spec: %v", err)
	}
	pi, ok := oa.Paths["/users/{id}"]
	if !ok {
		t.Fatalf("expected path /users/{id} in regenerated spec")
	}
	if pi.Get == nil || pi.Get.OperationID != "getUser" || pi.Get.Summary != "Get User" || pi.Get.Description != "Gets a user by ID" {
		t.Errorf("unexpected GET operation: %+v", pi.Get)
	}
	if pi.Post == nil || pi.Post.OperationID != "createUser" {
		t.Errorf("unexpected POST operation: %+v", pi.Post)
	}
	if pi.Put == nil || pi.Put.OperationID != "putUsersId" {
		t.Errorf("unexpected PUT operation: %+v", pi.Put)
	}
}

func assertGeneratedModuleDocCoverage(t *testing.T, dir string) {
	t.Helper()
	fset := token.NewFileSet()
	err := filepath.Walk(dir, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return walkErr
		}
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse generated file %s: %v", p, err)
		}
		if f.Doc == nil || strings.TrimSpace(f.Doc.Text()) == "" {
			t.Errorf("generated file %s is missing package doc comment", p)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok == token.TYPE || d.Tok == token.VAR || d.Tok == token.CONST {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if d.Doc == nil && s.Doc == nil {
								t.Errorf("generated type %s in %s is missing doc comment", s.Name.Name, p)
							}
							if st, ok := s.Type.(*ast.StructType); ok && st.Fields != nil {
								for _, field := range st.Fields.List {
									if len(field.Names) > 0 && field.Doc == nil && field.Comment == nil {
										t.Errorf("generated struct field %s.%s in %s is missing doc comment", s.Name.Name, field.Names[0].Name, p)
									}
								}
							} else if iface, ok := s.Type.(*ast.InterfaceType); ok && iface.Methods != nil {
								for _, method := range iface.Methods.List {
									if len(method.Names) > 0 && method.Doc == nil && method.Comment == nil {
										t.Errorf("generated interface method %s.%s in %s is missing doc comment", s.Name.Name, method.Names[0].Name, p)
									}
								}
							}
						case *ast.ValueSpec:
							if len(s.Names) > 0 && d.Doc == nil && s.Doc == nil {
								t.Errorf("generated value %s in %s is missing doc comment", s.Names[0].Name, p)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Doc == nil {
					t.Errorf("generated func %s in %s is missing doc comment", d.Name.Name, p)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
}

func TestCLIGithubOrgTopologyAndMCP(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.json")
	specContent := `{
		"openapi": "3.2.0",
		"info": {"title": "Petstore", "version": "1.0.0"},
		"paths": {
			"/pets": {
				"get": {"operationId": "listPets", "summary": "List pets"}
			}
		},
		"components": {
			"schemas": {
				"Pet": {
					"type": "object",
					"properties": {"id": {"type": "string"}, "name": {"type": "string"}}
				}
			}
		}
	}`
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	arazzoPath := filepath.Join(dir, "arazzo.json")
	arazzoContent := `{
		"arazzo": "1.1.0",
		"info": {"title": "Petstore Workflow", "version": "1.0.0"},
		"sourceDescriptions": [{"name": "petstore", "url": "./openapi.json", "type": "openapi"}],
		"workflows": [{"workflowId": "adoptPet", "summary": "Adopt a pet", "steps": [{"stepId": "step1", "operationId": "listPets"}]}]
	}`
	if err := os.WriteFile(arazzoPath, []byte(arazzoContent), 0644); err != nil {
		t.Fatalf("write arazzo: %v", err)
	}

	for _, target := range []string{"to_sdk", "to_sdk_cli", "to_server"} {
		oaOut := filepath.Join(dir, "oa_"+target)
		if err := run([]string{
			"from_openapi", target,
			"-i", specPath,
			"-o", oaOut,
			"--github-org", "acme",
			"--topology", "monorepo",
			"--tests",
			"--mcp",
		}); err != nil {
			t.Fatalf("from_openapi %s failed: %v", target, err)
		}
		goModBytes, err := os.ReadFile(filepath.Join(oaOut, "go.mod"))
		if err != nil || !strings.Contains(string(goModBytes), "module github.com/acme/acme-sdk/go") {
			t.Errorf("expected monorepo module path in %s/go.mod, got: %s", target, string(goModBytes))
		}
		assertGeneratedModuleDocCoverage(t, oaOut)

		azOut := filepath.Join(dir, "az_"+target)
		if err := run([]string{
			"from_arazzo", target,
			"-i", arazzoPath,
			"-o", azOut,
			"--github-org", "acme",
			"--repo-name", "custom-sdk",
			"--topology", "polyrepo",
			"--tests",
			"--mcp",
		}); err != nil {
			t.Fatalf("from_arazzo %s failed: %v", target, err)
		}
		azModBytes, err := os.ReadFile(filepath.Join(azOut, "go.mod"))
		if err != nil || !strings.Contains(string(azModBytes), "module github.com/acme/custom-sdk") {
			t.Errorf("expected polyrepo module path in %s/go.mod, got: %s", target, string(azModBytes))
		}
		assertGeneratedModuleDocCoverage(t, azOut)
	}
}
