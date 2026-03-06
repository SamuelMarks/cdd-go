package schemas

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseSchema(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.TYPE,
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// A user model"},
			},
		},
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("User"),
				Type: &dst.StructType{
					Fields: &dst.FieldList{
						List: []*dst.Field{
							{
								Names: []*dst.Ident{dst.NewIdent("Id")},
								Type:  dst.NewIdent("string"),
								Tag:   &dst.BasicLit{Kind: token.STRING, Value: "`json:\"id\"`"},
								Decs: dst.FieldDecorations{
									NodeDecs: dst.NodeDecs{
										Start: dst.Decorations{"// UUID"},
									},
								},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("Age")},
								Type:  dst.NewIdent("int"),
							},
							{
								Names: []*dst.Ident{dst.NewIdent("IsActive")},
								Type:  dst.NewIdent("bool"),
							},
							{
								Names: []*dst.Ident{dst.NewIdent("Score")},
								Type:  dst.NewIdent("float64"),
							},
							{
								Names: []*dst.Ident{dst.NewIdent("Profile")},
								Type:  &dst.StarExpr{X: dst.NewIdent("Profile")},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("Tags")},
								Type:  &dst.ArrayType{Elt: dst.NewIdent("string")},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("Metadata")},
								Type:  &dst.MapType{Value: dst.NewIdent("string")},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("CreatedAt")},
								Type:  &dst.SelectorExpr{X: dst.NewIdent("time"), Sel: dst.NewIdent("Time")},
							},
							{
								Names: []*dst.Ident{dst.NewIdent("AnyField")},
								Type:  dst.NewIdent("any"),
							},
							{
								// Empty names
								Type: dst.NewIdent("Embedded"),
							},
						},
					},
				},
			},
		},
	}

	name, schema := Parse(decl)
	if name != "User" {
		t.Errorf("expected User")
	}
	if schema == nil {
		t.Fatalf("expected schema")
	}
	if schema.Type != "object" {
		t.Errorf("expected object")
	}
	if schema.Description != "A user model" {
		t.Errorf("expected description")
	}

	if p, ok := schema.Properties["id"]; !ok || p.Type != "string" || p.Description != "UUID" {
		t.Errorf("expected id")
	}
	if p, ok := schema.Properties["age"]; !ok || p.Type != "integer" {
		t.Errorf("expected age")
	}
	if p, ok := schema.Properties["isActive"]; !ok || p.Type != "boolean" {
		t.Errorf("expected isActive")
	}
	if p, ok := schema.Properties["score"]; !ok || p.Type != "number" {
		t.Errorf("expected score")
	}
	if p, ok := schema.Properties["profile"]; !ok || p.Ref != "#/components/schemas/Profile" {
		t.Errorf("expected profile ref")
	}
	if p, ok := schema.Properties["tags"]; !ok || p.Type != "array" || p.Items.Type != "string" {
		t.Errorf("expected array")
	}
	if p, ok := schema.Properties["metadata"]; !ok || p.Type != "object" || p.AdditionalProperties.Type != "string" {
		t.Errorf("expected object/map")
	}
	if p, ok := schema.Properties["createdAt"]; !ok || p.Type != "string" || p.Format != "date-time" {
		t.Errorf("expected time")
	}
	if p, ok := schema.Properties["anyField"]; !ok || p.Type != "" {
		t.Errorf("expected any")
	}
}

func TestParseSchemaMapType(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent("StringMap"),
				Type: &dst.MapType{Value: dst.NewIdent("string")},
			},
		},
	}

	name, schema := Parse(decl)
	if name != "StringMap" {
		t.Errorf("expected StringMap")
	}
	if schema.Type != "object" || schema.AdditionalProperties.Type != "string" {
		t.Errorf("expected object with additional string properties")
	}
}

func TestParseSchemaNilAndErrors(t *testing.T) {
	n, s := Parse(nil)
	if n != "" || s != nil {
		t.Errorf("expected nil")
	}

	n, s = Parse(&dst.GenDecl{Tok: token.VAR})
	if n != "" || s != nil {
		t.Errorf("expected nil for VAR")
	}

	n, s = Parse(&dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{&dst.ValueSpec{}}})
	if n != "" || s != nil {
		t.Errorf("expected nil for non-TypeSpec")
	}
}

func TestParseTypeNilAndUnknown(t *testing.T) {
	s := ParseType(nil)
	if s.Type != "" {
		t.Errorf("expected empty")
	}

	s = ParseType(&dst.BasicLit{})
	if s.Type != "" {
		t.Errorf("expected empty")
	}
}
