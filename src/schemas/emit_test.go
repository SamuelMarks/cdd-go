package schemas

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitSchema(t *testing.T) {
	schema := &openapi.Schema{
		Type:        "object",
		Description: "A user model",
		Properties: map[string]openapi.Schema{
			"id":       {Type: "string", Description: "UUID"},
			"age":      {Type: "integer"},
			"isActive": {Type: "boolean"},
			"score":    {Type: "number"},
			"profile":  {Ref: "#/components/schemas/Profile"},
			"tags":     {Type: "array", Items: &openapi.Schema{Type: "string"}},
			"metadata": {Type: "object", AdditionalProperties: &openapi.Schema{Type: "string"}},
			"unknown":  {Type: "something_else"},
		},
	}

	decl, err := Emit("user", schema)
	if err != nil || decl == nil {
		t.Fatalf("expected decl, got err=%v", err)
	}

	file := &dst.File{
		Name:  dst.NewIdent("schemas"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err = restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.Join(strings.Fields(buf.String()), " ")

	if !strings.Contains(out, "// A user model") {
		t.Errorf("expected Description")
	}
	if !strings.Contains(out, "type User struct {") {
		t.Errorf("expected type User struct, got %s", out)
	}
	if !strings.Contains(out, "Id string `"+`json:"id,omitempty"`+"`") {
		t.Errorf("expected id")
	}
	if !strings.Contains(out, "// UUID") {
		t.Errorf("expected id desc")
	}
	if !strings.Contains(out, "Age int `"+`json:"age,omitempty"`+"`") {
		t.Errorf("expected age")
	}
	if !strings.Contains(out, "IsActive bool `"+`json:"isActive,omitempty"`+"`") {
		t.Errorf("expected isActive")
	}
	if !strings.Contains(out, "Score float64 `"+`json:"score,omitempty"`+"`") {
		t.Errorf("expected score")
	}
	if !strings.Contains(out, "Profile Profile `"+`json:"profile,omitempty"`+"`") {
		t.Errorf("expected profile")
	}
	if !strings.Contains(out, "Tags []string `"+`json:"tags,omitempty"`+"`") {
		t.Errorf("expected tags")
	}
	if !strings.Contains(out, "Metadata map[string]string `"+`json:"metadata,omitempty"`+"`") {
		t.Errorf("expected metadata")
	}
	if !strings.Contains(out, "Unknown interface{} `"+`json:"unknown,omitempty"`+"`") {
		t.Errorf("expected unknown fallback")
	}
}

func TestEmitSchemaNilAndNonObject(t *testing.T) {
	if decl, err := Emit("t", nil); err != nil || decl != nil {
		t.Errorf("expected nil, nil for nil schema")
	}
	if decl, err := Emit("t", &openapi.Schema{Type: "string"}); err != nil || decl == nil {
		t.Errorf("expected non-nil alias decl for string schema, got %v, %v", decl, err)
	}
	if _, err := Emit("", &openapi.Schema{Type: "object"}); err == nil {
		t.Errorf("expected error for empty name")
	}
	if _, err := Emit("t", &openapi.Schema{Type: "invalid-type"}); err == nil {
		t.Errorf("expected error for invalid schema type")
	}
}

func TestEmitTypeNil(t *testing.T) {
	expr := EmitType(nil)
	if ident, ok := expr.(*dst.Ident); !ok || ident.Name != "interface{}" {
		t.Errorf("expected interface{}")
	}
}

func TestEmitTypeObjectNoAdditional(t *testing.T) {
	expr := EmitType(&openapi.Schema{Type: "object"})
	if ident, ok := expr.(*dst.Ident); !ok || ident.Name != "interface{}" {
		t.Errorf("expected interface{}")
	}
}
