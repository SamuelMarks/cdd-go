package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComponentsBasic(t *testing.T) {
	dir, err := os.MkdirTemp("", "cdd-go-test-comp-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	specPath := filepath.Join(dir, "openapi.json")
	outDir := filepath.Join(dir, "generated")
	os.MkdirAll(outDir, 0755)

	specContent := `{
		"openapi": "3.2.0",
		"components": {
			"securitySchemes": { "basic": { "type": "http", "scheme": "basic" } },
			"parameters": { "limit": {"name": "limit", "in": "query", "schema": {"type": "integer"}} },
			"schemas": { "User": { "type": "object", "properties": { "id": {"type": "string"} } } },
			"headers": { "RateLimit": {"description": "Rate limit", "schema": {"type": "integer"}} },
			"requestBodies": {
				"UserBody": {
					"description": "A user",
					"content": {
						"application/json": {"schema": {"$ref": "#/components/schemas/User"}}
					}
				}
			},
			"responses": {
				"UserResp": {
					"description": "A user",
					"content": {
						"application/json": {"schema": {"$ref": "#/components/schemas/User"}}
					}
				}
			}
		}
	}`
	ioutil.WriteFile(specPath, []byte(specContent), 0644)

	err = run([]string{"from_openapi", "to_sdk", "-i", specPath, "-o", outDir})
	if err != nil {
		t.Fatalf("failed to generate: %v", err)
	}

	compCode, err := ioutil.ReadFile(filepath.Join(outDir, "components.go"))
	if err != nil {
		t.Fatalf("failed to read components.go: %v", err)
	}
	if !strings.Contains(string(compCode), "SecuritySchemeBasic") {
		t.Errorf("missing scheme")
	}
	if !strings.Contains(string(compCode), "type User struct") {
		t.Errorf("missing schema")
	}
	if !strings.Contains(string(compCode), "HeaderRateLimit") {
		t.Errorf("missing header")
	}
	if !strings.Contains(string(compCode), "RequestBodyUserBody") {
		t.Errorf("missing req body")
	}
	if !strings.Contains(string(compCode), "ResponseUserResp") {
		t.Errorf("missing resp")
	}

	outSpec := filepath.Join(dir, "openapi_regen.json")
	err = run([]string{"to_openapi", "-i", outDir, "-o", outSpec})
	if err != nil {
		t.Fatalf("failed to regenerate: %v", err)
	}

	regen, err := ioutil.ReadFile(outSpec)
	if err != nil {
		t.Fatalf("failed to read regenerated spec: %v", err)
	}
	if !strings.Contains(string(regen), `"basic"`) {
		t.Errorf("missing basic scheme in regen")
	}
	if !strings.Contains(string(regen), `"User"`) {
		t.Errorf("missing user schema in regen")
	}
	if !strings.Contains(string(regen), `"userBody"`) {
		t.Errorf("missing body in regen")
	}
	if !strings.Contains(string(regen), `"userResp"`) {
		t.Errorf("missing resp in regen")
	}
	if !strings.Contains(string(regen), `"rateLimit"`) {
		t.Errorf("missing header in regen")
	}
}
