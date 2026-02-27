package classes

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseType(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("User"),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("ID")},
						Type:  dst.NewIdent("string"),
						Tag: &dst.BasicLit{
							Kind:  token.STRING,
							Value: "`json:\"id\"`",
						},
						Decs: dst.FieldDecorations{
							NodeDecs: dst.NodeDecs{
								Start: dst.Decorations{"// User ID"},
							},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("Name")},
						Type:  dst.NewIdent("string"),
						Tag: &dst.BasicLit{
							Kind:  token.STRING,
							Value: "`json:\"name,omitempty\"`",
						},
					},
					{
						// embedded field
						Type: dst.NewIdent("Embedded"),
					},
				},
			},
		},
		Decs: dst.TypeSpecDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{"// User Profile"},
			},
		},
	}

	schema, err := ParseType(ts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.Description != "User Profile" {
		t.Errorf("expected User Profile, got %s", schema.Description)
	}

	if schema.Type != "object" {
		t.Errorf("expected object, got %s", schema.Type)
	}

	if len(schema.Properties) != 2 {
		t.Errorf("expected 2 property, got %d", len(schema.Properties))
	}

	prop := schema.Properties["id"]
	if prop.Type != "string" {
		t.Errorf("expected string, got %s", prop.Type)
	}

	if prop.Description != "User ID" {
		t.Errorf("expected User ID, got %s", prop.Description)
	}

	propName := schema.Properties["name"]
	if propName.Type != "string" {
		t.Errorf("expected string, got %s", propName.Type)
	}
}

func TestParseTypeArray(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("StringList"),
		Type: &dst.ArrayType{
			Elt: dst.NewIdent("string"),
		},
	}
	schema, err := ParseType(ts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if schema.Type != "array" {
		t.Errorf("expected array")
	}
	if schema.Items == nil || schema.Items.Type != "string" {
		t.Errorf("expected string items")
	}
}

func TestParseTypeScalar(t *testing.T) {
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("StringAlias"),
		Type: dst.NewIdent("string"),
	}
	schema, err := ParseType(ts)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if schema.Type != "string" {
		t.Errorf("expected string")
	}
}

func TestParseExpr(t *testing.T) {
	exprs := []struct {
		expr     dst.Expr
		expected string
		isRef    bool
	}{
		{dst.NewIdent("string"), "string", false},
		{dst.NewIdent("int"), "integer", false},
		{dst.NewIdent("int8"), "integer", false},
		{dst.NewIdent("int16"), "integer", false},
		{dst.NewIdent("int32"), "integer", false},
		{dst.NewIdent("int64"), "integer", false},
		{dst.NewIdent("uint"), "integer", false},
		{dst.NewIdent("uint8"), "integer", false},
		{dst.NewIdent("uint16"), "integer", false},
		{dst.NewIdent("uint32"), "integer", false},
		{dst.NewIdent("uint64"), "integer", false},
		{dst.NewIdent("float32"), "number", false},
		{dst.NewIdent("float64"), "number", false},
		{dst.NewIdent("bool"), "boolean", false},
		{dst.NewIdent("interface{}"), "", false},
		{dst.NewIdent("CustomType"), "#/components/schemas/CustomType", true},
		{&dst.StarExpr{X: dst.NewIdent("string")}, "string", false},
		{&dst.ArrayType{Elt: dst.NewIdent("string")}, "array", false},
		{&dst.StructType{}, "object", false},
	}

	for _, tc := range exprs {
		schema, err := ParseExpr(tc.expr)
		if err != nil {
			t.Fatal(err)
		}
		if tc.isRef {
			if schema.Ref != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, schema.Ref)
			}
		} else {
			if schema.Type != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, schema.Type)
			}
		}
	}
}

func TestParseTypeNil(t *testing.T) {
	_, err := ParseType(nil)
	if err == nil {
		t.Errorf("expected error for nil TypeSpec")
	}
}

func TestParseExprErrors(t *testing.T) {
	_, err := ParseExpr(nil)
	if err == nil {
		t.Errorf("expected error")
	}

	_, err = ParseExpr(&dst.ChanType{})
	if err == nil {
		t.Errorf("expected error for unsupported chan type")
	}
}

func TestParseTypeWithErrors(t *testing.T) {
	// struct field error
	ts := &dst.TypeSpec{
		Name: dst.NewIdent("StructErr"),
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("F")},
						Type:  nil,
					},
				},
			},
		},
	}
	_, err := ParseType(ts)
	if err == nil {
		t.Errorf("expected error")
	}

	// array element error
	tsArr := &dst.TypeSpec{
		Name: dst.NewIdent("ArrErr"),
		Type: &dst.ArrayType{Elt: nil},
	}
	_, err = ParseType(tsArr)
	if err == nil {
		t.Errorf("expected error")
	}

	// default error
	tsDef := &dst.TypeSpec{
		Name: dst.NewIdent("DefErr"),
		Type: &dst.ChanType{},
	}
	_, err = ParseType(tsDef)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseExprArrayErr(t *testing.T) {
	_, err := ParseExpr(&dst.ArrayType{Elt: nil})
	if err == nil {
		t.Errorf("expected error")
	}
}
