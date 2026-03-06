package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICommandsEndToEnd(t *testing.T) {
	dir, err := os.MkdirTemp("", "cdd-go-test-cli-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	specPath := filepath.Join(dir, "openapi.json")
	outDir := filepath.Join(dir, "generated")
	os.MkdirAll(outDir, 0755)

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
	ioutil.WriteFile(specPath, []byte(specContent), 0644)

	err = run([]string{"from_openapi", "to_sdk_cli", "-i", specPath, "-o", outDir})
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	cliCode, err := ioutil.ReadFile(filepath.Join(outDir, "sdk_cli.go"))
	if err != nil {
		t.Fatalf("failed to read cli.go: %v", err)
	}
	sCliCode := string(cliCode)

	if !strings.Contains(sCliCode, `var GetUserCmd = &cobra.Command`) {
		t.Errorf("missing GetUserCmd")
	}

	outSpec := filepath.Join(dir, "openapi_regen.json")
	err = run([]string{"to_openapi", "-i", outDir, "-o", outSpec})
	if err != nil {
		t.Fatalf("failed to regenerate: %v", err)
	}
}
