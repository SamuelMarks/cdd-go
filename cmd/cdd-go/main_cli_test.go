package main

import (
	"encoding/json"
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
