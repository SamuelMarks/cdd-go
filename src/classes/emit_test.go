package classes

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitType(t *testing.T) {
	schema := &openapi.Schema{
		Type:        "object",
		Description: "User profile",
		Properties: map[string]openapi.Schema{
			"id": {
				Type:        "string",
				Description: "Unique identifier",
			},
			"age": {
				Type: "integer",
			},
			"friends": {
				Type: "array",
				Items: &openapi.Schema{
					Type: "string",
				},
			},
			"settings": {
				Ref: "#/components/schemas/Settings",
			},
		},
	}

	ts, err := EmitType("User", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ts.Name.Name != "User" {
		t.Errorf("expected User, got %s", ts.Name.Name)
	}

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}

	file := &dst.File{
		Name:  dst.NewIdent("classes"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err = restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "  ", " ") // collapse spaces for easier assertion
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "type // User profile") && !strings.Contains(out, "User struct") {
		t.Errorf("expected User struct, got %s", out)
	}
	if !strings.Contains(out, "// Unique identifier") {
		t.Errorf("expected property description")
	}
	if !strings.Contains(out, "Age int `json:\"age\" gorm:\"column:age\"`") {
		t.Errorf("expected int age field, got %s", out)
	}
	if !strings.Contains(out, "Friends []string `json:\"friends\" gorm:\"column:friends\"`") {
		t.Errorf("expected string array friends field, got %s", out)
	}
	if !strings.Contains(out, "Settings Settings `json:\"settings\" gorm:\"column:settings\"`") {
		t.Errorf("expected settings field, got %s", out)
	}
}

func TestEmitTypeArray(t *testing.T) {
	schema := &openapi.Schema{
		Type: "array",
		Items: &openapi.Schema{
			Type: "string",
		},
	}
	ts, err := EmitType("StringList", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Name.Name != "StringList" {
		t.Errorf("expected StringList")
	}
	_, ok := ts.Type.(*dst.ArrayType)
	if !ok {
		t.Errorf("expected array type")
	}
}

func TestEmitTypeScalar(t *testing.T) {
	schema := &openapi.Schema{
		Type: "string",
	}
	ts, err := EmitType("StringAlias", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Name.Name != "StringAlias" {
		t.Errorf("expected StringAlias")
	}
	ident, ok := ts.Type.(*dst.Ident)
	if !ok || ident.Name != "string" {
		t.Errorf("expected string ident")
	}
}

func TestEmitTypeExpr(t *testing.T) {
	if expr := EmitTypeExpr(nil); expr.(*dst.Ident).Name != "interface{}" {
		t.Errorf("expected interface{}")
	}

	numExpr := EmitTypeExpr(&openapi.Schema{Type: "number"})
	if numExpr.(*dst.Ident).Name != "float64" {
		t.Errorf("expected float64")
	}

	boolExpr := EmitTypeExpr(&openapi.Schema{Type: "boolean"})
	if boolExpr.(*dst.Ident).Name != "bool" {
		t.Errorf("expected bool")
	}

	strExpr := EmitTypeExpr(&openapi.Schema{Type: "string"})
	if strExpr.(*dst.Ident).Name != "string" {
		t.Errorf("expected string")
	}

	intExpr := EmitTypeExpr(&openapi.Schema{Type: "integer"})
	if intExpr.(*dst.Ident).Name != "int" {
		t.Errorf("expected int")
	}

	arrExpr := EmitTypeExpr(&openapi.Schema{Type: "array", Items: &openapi.Schema{Type: "string"}})
	if _, ok := arrExpr.(*dst.ArrayType); !ok {
		t.Errorf("expected ArrayType")
	}

	unknownExpr := EmitTypeExpr(&openapi.Schema{Type: "unknown"})
	if unknownExpr.(*dst.Ident).Name != "interface{}" {
		t.Errorf("expected interface{}")
	}
}

func TestEmitTypeNil(t *testing.T) {
	_, err := EmitType("User", nil)
	if err == nil {
		t.Errorf("expected error for nil schema")
	}
	_, err = EmitType("", &openapi.Schema{Type: "string"})
	if err == nil {
		t.Errorf("expected error for empty schema name")
	}
	_, err = EmitType("User", &openapi.Schema{Type: "invalid-type"})
	if err == nil {
		t.Errorf("expected error for unsupported schema type")
	}
}

func TestEmitTypeMap(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
		AdditionalProperties: &openapi.Schema{
			Type: "string",
		},
	}
	ts, err := EmitType("MapType", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mt, ok := ts.Type.(*dst.MapType)
	if !ok {
		t.Errorf("expected map type, got %T", ts.Type)
	}
	if mt.Key.(*dst.Ident).Name != "string" {
		t.Errorf("expected string key")
	}
	if mt.Value.(*dst.Ident).Name != "string" {
		t.Errorf("expected string value")
	}
}

func TestEmitTypeExprMap(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
		AdditionalProperties: &openapi.Schema{
			Type: "integer",
		},
	}
	expr := EmitTypeExpr(schema)
	mt, ok := expr.(*dst.MapType)
	if !ok {
		t.Errorf("expected map type, got %T", expr)
	}
	if mt.Key.(*dst.Ident).Name != "string" {
		t.Errorf("expected string key")
	}
	if mt.Value.(*dst.Ident).Name != "int" {
		t.Errorf("expected int value")
	}
}

func TestEmitTypeExprObjectFallthrough(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
	}
	expr := EmitTypeExpr(schema)
	id, ok := expr.(*dst.Ident)
	if !ok || id.Name != "interface{}" {
		t.Errorf("expected interface{}")
	}
}

func TestEmitTypeWithOptions(t *testing.T) {
	objSchema := &openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"id":    {Type: "string"},
			"i32":   {Type: "integer", Format: "int32"},
			"i64":   {Type: "integer", Format: "int64"},
			"f32":   {Type: "number", Format: "float"},
			"inner": {Type: "object", Properties: map[string]openapi.Schema{"code": {Type: "string"}}},
		},
	}
	tsOmit, err := EmitTypeWithOptions("User", objSchema, EmitOptions{
		IncludeGormTags:         false,
		OmitEmpty:               true,
		UsePascalCaseFieldNames: true,
	})
	if err != nil || tsOmit == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tsPlain, err := EmitTypeWithOptions("User", objSchema, EmitOptions{
		IncludeGormTags: false,
		OmitEmpty:       false,
	})
	if err != nil || tsPlain == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var buf bytes.Buffer
	if err := decorator.NewRestorer().Fprint(&buf, &dst.File{
		Name:  dst.NewIdent("models"),
		Decls: []dst.Decl{&dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{tsPlain}}},
	}); err != nil {
		t.Fatalf("failed to print plain struct: %v", err)
	}
	plainCode := buf.String()
	if !strings.Contains(plainCode, "`json:\"id\"`") || strings.Contains(plainCode, "gorm:") || strings.Contains(plainCode, "omitempty") {
		t.Errorf("expected plain json tag without gorm or omitempty, got:\n%s", plainCode)
	}
	if !strings.Contains(plainCode, "int32") || !strings.Contains(plainCode, "int64") || !strings.Contains(plainCode, "float32") || !strings.Contains(plainCode, "Code string") {
		t.Errorf("expected int32, int64, float32, and inline struct field Code string, got:\n%s", plainCode)
	}

	discXMLSchema := &openapi.Schema{
		Type:        "object",
		Description: "Polymorphic Pet",
		Discriminator: &openapi.Discriminator{
			PropertyName:   "petType",
			DefaultMapping: "#/components/schemas/DefaultPet",
		},
		XML: &openapi.XML{
			Name:     "pet",
			NodeType: "element",
		},
		Properties: map[string]openapi.Schema{
			"id":   {Type: "string", XML: &openapi.XML{NodeType: "attribute"}},
			"raw":  {Type: "string", XML: &openapi.XML{Name: "rawData", NodeType: "cdata"}},
			"body": {Type: "string", XML: &openapi.XML{Name: "bodyText", NodeType: "text"}},
		},
	}
	tsDisc, err := EmitType("Pet", discXMLSchema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsedBack, err := ParseType(tsDisc)
	if err != nil {
		t.Fatalf("unexpected ParseType error: %v", err)
	}
	if parsedBack.Discriminator == nil || parsedBack.Discriminator.DefaultMapping != "#/components/schemas/DefaultPet" {
		t.Errorf("expected DefaultMapping roundtrip, got %+v", parsedBack.Discriminator)
	}
	if parsedBack.XML == nil || parsedBack.XML.NodeType != "element" {
		t.Errorf("expected XML NodeType roundtrip, got %+v", parsedBack.XML)
	}
	if parsedBack.Properties["id"].XML == nil || parsedBack.Properties["id"].XML.NodeType != "attribute" {
		t.Errorf("expected id XML attribute roundtrip, got %+v", parsedBack.Properties["id"].XML)
	}
	if parsedBack.Properties["raw"].XML == nil || parsedBack.Properties["raw"].XML.NodeType != "cdata" {
		t.Errorf("expected raw XML cdata roundtrip, got %+v", parsedBack.Properties["raw"].XML)
	}
	if parsedBack.Properties["body"].XML == nil || parsedBack.Properties["body"].XML.NodeType != "text" {
		t.Errorf("expected body XML text roundtrip, got %+v", parsedBack.Properties["body"].XML)
	}
}
