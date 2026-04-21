package requestbodies

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitRequestBody(t *testing.T) {
	rb := &openapi.RequestBody{
		Description: "A test payload",
		Required:    true,
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Ref: "#/components/schemas/User"},
			},
		},
	}

	field := Emit(rb)
	if field == nil {
		t.Fatalf("expected field, got nil")
	}

	decl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("Dummy"),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{field},
					},
				},
			},
		},
	}

	file := &dst.File{
		Name:  dst.NewIdent("requestbodies"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "// A test payload") {
		t.Errorf("expected Description")
	}
	if !strings.Contains(out, "// Required: true") {
		t.Errorf("expected Required")
	}
	if !strings.Contains(out, "body User") {
		t.Errorf("expected body User, got %s", out)
	}
}

func TestEmitRequestBodyNil(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestEmitRequestBodyEmpty(t *testing.T) {
	rb := &openapi.RequestBody{}
	field := Emit(rb)
	if len(field.Decs.Start) > 0 {
		t.Errorf("expected no docstrings")
	}
}

func TestEmitRequestBodyArray(t *testing.T) {
	rb := &openapi.RequestBody{
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/User"}},
			},
		},
	}
	f := Emit(rb)
	if _, ok := f.Type.(*dst.ArrayType); !ok {
		t.Errorf("expected array")
	}
}

func TestEmitRequestBodyObject(t *testing.T) {
	rb := &openapi.RequestBody{
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{Type: "object"},
			},
		},
	}
	f := Emit(rb)
	if _, ok := f.Type.(*dst.MapType); !ok {
		t.Errorf("expected map")
	}
}
