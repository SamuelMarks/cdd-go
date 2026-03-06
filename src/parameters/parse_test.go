package parameters

import (
	"github.com/dave/dst"
	"testing"
)

func TestParseParameter(t *testing.T) {
	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("limit")},
		Type:  dst.NewIdent("int"),
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// A limit param",
					"// In: header",
					"// Required: true",
					"// Deprecated",
				},
			},
		},
	}

	p := Parse(field)
	if p == nil {
		t.Fatalf("expected parameter")
	}
	if p.Name != "limit" {
		t.Errorf("expected limit")
	}
	if p.Description != "A limit param" {
		t.Errorf("expected description")
	}
	if p.In != "header" {
		t.Errorf("expected in header, got %s", p.In)
	}
	if !p.Required {
		t.Errorf("expected required")
	}
	if !p.Deprecated {
		t.Errorf("expected deprecated")
	}
	if p.Schema.Type != "integer" {
		t.Errorf("expected integer schema type")
	}
}

func TestParseParameterId(t *testing.T) {
	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("userId")},
		Type:  dst.NewIdent("string"),
	}

	p := Parse(field)
	if p.In != "path" {
		t.Errorf("expected path for id field")
	}
	if !p.Required {
		t.Errorf("expected required for path field")
	}
}

func TestParseParameterTypes(t *testing.T) {
	cases := []struct {
		IdentName    string
		ExpectedType string
		ExpectedRef  string
	}{
		{"bool", "boolean", ""},
		{"float64", "number", ""},
		{"string", "string", ""},
		{"CustomParam", "", "#/components/schemas/CustomParam"},
	}
	for _, tc := range cases {
		field := &dst.Field{
			Names: []*dst.Ident{dst.NewIdent("test")},
			Type:  dst.NewIdent(tc.IdentName),
		}
		p := Parse(field)
		if p == nil {
			t.Fatalf("expected parameter for %s", tc.IdentName)
		}
		if p.Schema.Type != tc.ExpectedType {
			t.Errorf("expected type %s, got %s", tc.ExpectedType, p.Schema.Type)
		}
		if p.Schema.Ref != tc.ExpectedRef {
			t.Errorf("expected ref %s, got %s", tc.ExpectedRef, p.Schema.Ref)
		}
	}
}

func TestParseParameterNil(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
	if Parse(&dst.Field{}) != nil {
		t.Errorf("expected nil for empty names")
	}
}
