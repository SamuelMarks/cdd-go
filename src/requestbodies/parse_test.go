package requestbodies

import (
	"github.com/dave/dst"
	"testing"
)

func TestParseRequestBody(t *testing.T) {
	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("body")},
		Type:  dst.NewIdent("User"),
		Decs: dst.FieldDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// User payload",
					"// Required: true",
				},
			},
		},
	}

	rb := Parse(field)
	if rb == nil {
		t.Fatalf("expected request body")
	}
	if rb.Description != "User payload" {
		t.Errorf("expected description")
	}
	if !rb.Required {
		t.Errorf("expected required")
	}
	mt, ok := rb.Content["application/json"]
	if !ok || mt.Schema == nil {
		t.Fatalf("expected json schema")
	}
	if mt.Schema.Ref != "#/components/schemas/User" {
		t.Errorf("expected User ref")
	}
}

func TestParseRequestBodyPointerAndArray(t *testing.T) {
	field1 := &dst.Field{
		Type: &dst.StarExpr{X: dst.NewIdent("User")},
	}
	rb1 := Parse(field1)
	if rb1.Content["application/json"].Schema.Ref != "#/components/schemas/User" {
		t.Errorf("expected star User ref")
	}

	field2 := &dst.Field{
		Type: &dst.ArrayType{Elt: dst.NewIdent("User")},
	}
	rb2 := Parse(field2)
	if rb2.Content["application/json"].Schema.Type != "array" {
		t.Errorf("expected array type")
	}
	if rb2.Content["application/json"].Schema.Items.Ref != "#/components/schemas/User" {
		t.Errorf("expected User item ref")
	}
}

func TestParseRequestBodyNil(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestParseRequestBodyEmpty(t *testing.T) {
	field := &dst.Field{
		Type: &dst.MapType{}, // Fallback type
	}
	rb := Parse(field)
	if rb.Content["application/json"].Schema.Type != "object" {
		t.Errorf("expected fallback object type")
	}
}

func TestParseRequestBodyAny(t *testing.T) {
	field := &dst.Field{
		Type: dst.NewIdent("any"),
	}
	rb := Parse(field)
	if rb.Content["application/json"].Schema.Type != "object" {
		t.Errorf("expected object")
	}
}

func TestParseRequestBodyMap(t *testing.T) {
	field := &dst.Field{
		Type: &dst.MapType{},
	}
	rb := Parse(field)
	if rb.Content["application/json"].Schema.Type != "object" {
		t.Errorf("expected object for fallback")
	}
}

func TestParseRequestBodyString(t *testing.T) {
	field := &dst.Field{
		Type: dst.NewIdent("string"),
	}
	rb := Parse(field)
	if rb.Content["application/json"].Schema.Type != "string" {
		t.Errorf("expected string")
	}
}
