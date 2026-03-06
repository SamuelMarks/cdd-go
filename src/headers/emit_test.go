package headers

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitHeader(t *testing.T) {
	header := &openapi.Header{
		Description: "A header desc",
		Required:    true,
		Deprecated:  true,
		Schema:      &openapi.Schema{Type: "integer"},
	}

	field := Emit("X-Rate-Limit", header)
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
		Name:  dst.NewIdent("headers"),
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

	if !strings.Contains(out, "// A header desc") {
		t.Errorf("expected Description")
	}
	if !strings.Contains(out, "// Required: true") {
		t.Errorf("expected Required")
	}
	if !strings.Contains(out, "// Deprecated") {
		t.Errorf("expected Deprecated")
	}
	if !strings.Contains(out, "X-Rate-Limit int") {
		t.Errorf("expected X-Rate-Limit int, got %s", out)
	}
}

func TestEmitHeaderTypes(t *testing.T) {
	cases := []struct {
		Schema       *openapi.Schema
		ExpectedType string
	}{
		{&openapi.Schema{Type: "boolean"}, "bool"},
		{&openapi.Schema{Type: "number"}, "float64"},
		{&openapi.Schema{Ref: "#/components/schemas/MyHeader"}, "MyHeader"},
		{nil, "string"},
	}
	for _, tc := range cases {
		header := &openapi.Header{Schema: tc.Schema}
		f := Emit("test", header)
		ident := f.Type.(*dst.Ident)
		if ident.Name != tc.ExpectedType {
			t.Errorf("expected %s, got %s", tc.ExpectedType, ident.Name)
		}
	}
}

func TestEmitHeaderNil(t *testing.T) {
	if Emit("nil", nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestEmitHeaderEmpty(t *testing.T) {
	header := &openapi.Header{}
	field := Emit("Empty", header)
	if len(field.Decs.Start) > 0 {
		t.Errorf("expected no docstrings")
	}
}
