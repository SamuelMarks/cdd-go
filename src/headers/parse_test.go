package headers

import (
	"github.com/dave/dst"
	"testing"
)

func TestParseHeader(t *testing.T) {
	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("X-Test-Id")},
		Type:  dst.NewIdent("int"),
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// A desc",
					"// Required: true",
					"// Deprecated",
				},
			},
		},
	}

	header := Parse(field)
	if header == nil {
		t.Fatalf("expected header")
	}
	if header.Description != "A desc" {
		t.Errorf("expected description")
	}
	if !header.Required {
		t.Errorf("expected required")
	}
	if !header.Deprecated {
		t.Errorf("expected deprecated")
	}
	if header.Schema.Type != "integer" {
		t.Errorf("expected integer schema type")
	}
}

func TestParseHeaderTypes(t *testing.T) {
	cases := []struct {
		IdentName    string
		ExpectedType string
		ExpectedRef  string
	}{
		{"bool", "boolean", ""},
		{"float64", "number", ""},
		{"string", "string", ""},
		{"CustomHeader", "", "#/components/schemas/CustomHeader"},
	}
	for _, tc := range cases {
		field := &dst.Field{
			Names: []*dst.Ident{dst.NewIdent("test")},
			Type:  dst.NewIdent(tc.IdentName),
		}
		header := Parse(field)
		if header == nil {
			t.Fatalf("expected header for %s", tc.IdentName)
		}
		if header.Schema.Type != tc.ExpectedType {
			t.Errorf("expected type %s, got %s", tc.ExpectedType, header.Schema.Type)
		}
		if header.Schema.Ref != tc.ExpectedRef {
			t.Errorf("expected ref %s, got %s", tc.ExpectedRef, header.Schema.Ref)
		}
	}
}

func TestParseHeaderNil(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestParseHeaderEmpty(t *testing.T) {
	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("Empty")},
		Type:  &dst.StarExpr{}, // Non-ident
	}
	if Parse(field) != nil {
		t.Errorf("expected nil for non-ident without comments")
	}
}
