package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestComponentsBasic(t *testing.T) {
	dir, err := os.MkdirTemp("", "cdd-go-test-comp-*")
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
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("failed to write spec: %v", err)
	}

	err = run([]string{"from_openapi", "to_sdk", "-i", specPath, "-o", outDir})
	if err != nil {
		t.Fatalf("from_openapi to_sdk failed: %v", err)
	}

	compCode, err := os.ReadFile(filepath.Join(outDir, "models", "components.go"))
	if err != nil {
		t.Fatalf("failed to read components.go: %v", err)
	}
	if !strings.Contains(string(compCode), "SecuritySchemeBasic") {
		t.Errorf("expected SecuritySchemeBasic in components.go")
	}
	if !strings.Contains(string(compCode), "type User struct") {
		t.Errorf("expected type User struct in components.go")
	}
	if !strings.Contains(string(compCode), "HeaderRateLimit") {
		t.Errorf("expected HeaderRateLimit in components.go")
	}
	if !strings.Contains(string(compCode), "RequestBodyUserBody") {
		t.Errorf("expected RequestBodyUserBody in components.go")
	}
	if !strings.Contains(string(compCode), "ResponseUserResp") {
		t.Errorf("expected ResponseUserResp in components.go")
	}

	userModelCode, err := os.ReadFile(filepath.Join(outDir, "models", "user.go"))
	if err != nil {
		t.Fatalf("failed to read user.go: %v", err)
	}
	if !strings.Contains(string(userModelCode), "type User struct") {
		t.Errorf("expected type User struct in user.go")
	}

	outSpec := filepath.Join(dir, "openapi_regen.json")
	err = run([]string{"to_openapi", "-i", outDir, "-o", outSpec})
	if err != nil {
		t.Fatalf("to_openapi failed: %v", err)
	}

	regen, err := os.ReadFile(outSpec)
	if err != nil {
		t.Fatalf("failed to read regenerated spec: %v", err)
	}
	if !strings.Contains(string(regen), `"basic"`) {
		t.Errorf("expected basic security scheme in regen")
	}
	if !strings.Contains(string(regen), `"User"`) {
		t.Errorf("expected User schema in regen")
	}
	if !strings.Contains(string(regen), `"userBody"`) && !strings.Contains(string(regen), `"UserBody"`) {
		t.Errorf("expected userBody requestBody in regen")
	}
	if !strings.Contains(string(regen), `"userResp"`) && !strings.Contains(string(regen), `"UserResp"`) {
		t.Errorf("expected userResp response in regen")
	}
	if !strings.Contains(string(regen), `"rateLimit"`) && !strings.Contains(string(regen), `"RateLimit"`) {
		t.Errorf("expected rateLimit header in regen")
	}

	var oa openapi.OpenAPI
	if err := json.Unmarshal(regen, &oa); err != nil {
		t.Fatalf("failed to unmarshal regenerated spec: %v", err)
	}
	if oa.Components == nil {
		t.Fatalf("expected non-nil components")
	}
	if _, ok := oa.Components.SecuritySchemes["basic"]; !ok {
		t.Errorf("missing basic security scheme in parsed components")
	}
	if _, ok := oa.Components.Schemas["User"]; !ok {
		t.Errorf("missing User schema in parsed components")
	}
}
