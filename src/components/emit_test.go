package components

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitComponents(t *testing.T) {
	comp := &openapi.Components{
		SecuritySchemes: map[string]openapi.SecurityScheme{
			"basic": {Type: "http", Scheme: "basic"},
		},
		Parameters: map[string]openapi.Parameter{
			"limit": {Name: "limit", In: "query", Schema: &openapi.Schema{Type: "integer"}, Description: "limit param"},
		},
		Headers: map[string]openapi.Header{
			"RateLimit": {Description: "Rate limit", Schema: &openapi.Schema{Type: "integer"}},
		},
		RequestBodies: map[string]openapi.RequestBody{
			"UserBody": {
				Description: "A user",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/User"}},
				},
			},
		},
		Responses: map[string]openapi.Response{
			"UserResp": {
				Description: "A user",
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/User"}},
				},
			},
			"ComplexResp": {
				Headers: map[string]openapi.Header{"X-RateLimit": {Schema: &openapi.Schema{Type: "integer"}}},
				Content: map[string]openapi.MediaType{
					"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/User"}},
				},
			},
		},
		Schemas: map[string]openapi.Schema{
			"User": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"id": {Type: "string"},
				},
			},
		},
	}

	decls := Emit(comp)
	if len(decls) != 7 {
		t.Fatalf("expected 7 decls, got %d", len(decls))
	}

	file := &dst.File{
		Name:  dst.NewIdent("components"),
		Decls: decls,
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "\n", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "var SecuritySchemeBasic") {
		t.Errorf("missing scheme")
	}
	if !strings.Contains(out, "var ParamLimit int") {
		t.Errorf("missing param")
	}
	if !strings.Contains(out, "// limit param") {
		t.Errorf("missing param description")
	}
	if !strings.Contains(out, "var HeaderRateLimit int") {
		t.Errorf("missing header")
	}
	if !strings.Contains(out, "// Rate limit") {
		t.Errorf("missing header desc")
	}
	if !strings.Contains(out, "type RequestBodyUserBody User") {
		t.Errorf("missing req body")
	}
	if !strings.Contains(out, "// A user") {
		t.Errorf("missing req body desc")
	}
	if !strings.Contains(out, "type ResponseUserResp *User") {
		t.Errorf("missing basic resp")
	}
	if !strings.Contains(out, "type ResponseComplexResp struct { F0 *User F1 int }") {
		t.Errorf("missing complex resp, got %s", out)
	}
	if !strings.Contains(out, "type User struct") {
		t.Errorf("missing user schema, got %s", out)
	}
}

func TestEmitComponentsNil(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestToPascalCaseEmpty(t *testing.T) {
	if toPascalCase("") != "" {
		t.Errorf("expected empty")
	}
	if toPascalCase("Test") != "Test" {
		t.Errorf("expected Test")
	}
}
