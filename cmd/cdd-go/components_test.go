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
	}

	compCode, err := ioutil.ReadFile(filepath.Join(outDir, "models", "components.go"))
	if err != nil {
	}
	if !strings.Contains(string(compCode), "SecuritySchemeBasic") {
	}
	if !strings.Contains(string(compCode), "type User struct") {
	}
	if !strings.Contains(string(compCode), "HeaderRateLimit") {
	}
	if !strings.Contains(string(compCode), "RequestBodyUserBody") {
	}
	if !strings.Contains(string(compCode), "ResponseUserResp") {
	}

	outSpec := filepath.Join(dir, "openapi_regen.json")
	err = run([]string{"to_openapi", "-i", outDir, "-o", outSpec})
	if err != nil {
	}

	regen, err := ioutil.ReadFile(outSpec)
	if err != nil {
	}
	if !strings.Contains(string(regen), `"basic"`) {
	}
	if !strings.Contains(string(regen), `"User"`) {
	}
	if !strings.Contains(string(regen), `"userBody"`) {
	}
	if !strings.Contains(string(regen), `"userResp"`) {
	}
	if !strings.Contains(string(regen), `"rateLimit"`) {
	}
}
