package classes

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
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
	if !strings.Contains(out, "Age int `json:\"age\"`") {
		t.Errorf("expected int age field, got %s", out)
	}
	if !strings.Contains(out, "Friends []string `json:\"friends\"`") {
		t.Errorf("expected string array friends field, got %s", out)
	}
	if !strings.Contains(out, "Settings Settings `json:\"settings\"`") {
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

func TestExportedName(t *testing.T) {
	if exportedName("") != "" {
		t.Errorf("expected empty string")
	}
	if exportedName("id") != "Id" {
		t.Errorf("expected Id")
	}
}

func TestEmitTypeNil(t *testing.T) {
	_, err := EmitType("User", nil)
	if err == nil {
		t.Errorf("expected error for nil schema")
	}
}
