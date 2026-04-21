package parameters

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitParameter(t *testing.T) {
	p := openapi.Parameter{
		Name:        "limit",
		In:          "header",
		Description: "A limit param",
		Required:    true,
		Deprecated:  true,
		Schema:      &openapi.Schema{Type: "integer"},
	}

	field := Emit(p)
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
		Name:  dst.NewIdent("parameters"),
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

	if !strings.Contains(out, "// A limit param") {
		t.Errorf("expected Description")
	}
	if !strings.Contains(out, "// In: header") {
		t.Errorf("expected In header")
	}
	if !strings.Contains(out, "// Required: true") {
		t.Errorf("expected Required")
	}
	if !strings.Contains(out, "// Deprecated") {
		t.Errorf("expected Deprecated")
	}
	if !strings.Contains(out, "limit int") {
		t.Errorf("expected limit int, got %s", out)
	}
}

func TestEmitParameterTypes(t *testing.T) {
	cases := []struct {
		Schema       *openapi.Schema
		Ref          string
		ExpectedType string
	}{
		{&openapi.Schema{Type: "boolean"}, "", "bool"},
		{&openapi.Schema{Type: "number"}, "", "float64"},
		{&openapi.Schema{Ref: "#/components/schemas/MyParam"}, "", "MyParam"},
		{nil, "#/components/schemas/RefParam", "RefParam"},
		{nil, "", "string"},
	}
	for _, tc := range cases {
		p := openapi.Parameter{Name: "test", Schema: tc.Schema, Ref: tc.Ref}
		f := Emit(p)
		ident := f.Type.(*dst.Ident)
		if ident.Name != tc.ExpectedType {
			t.Errorf("expected %s, got %s", tc.ExpectedType, ident.Name)
		}
	}
}

func TestEmitParameterRefName(t *testing.T) {
	p := openapi.Parameter{Ref: "#/components/parameters/RefParamName"}
	f := Emit(p)
	if f.Names[0].Name != "RefParamName" {
		t.Errorf("expected RefParamName, got %s", f.Names[0].Name)
	}
}
