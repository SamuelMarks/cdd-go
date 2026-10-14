// Package schemas provides parsing and emitting of OpenAPI Schema objects to and from Go AST declarations.
package schemas

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Schema object into a struct representation.
// Parameter name is the schema name, and schema is the OpenAPI Schema definition.
func Emit(name string, schema *openapi.Schema) *dst.GenDecl {
	if schema == nil || schema.Type != "object" {
		return nil
	}

	ts, _ := classes.EmitTypeWithOptions(toPascalCase(name), schema, classes.EmitOptions{
		IncludeGormTags:         false,
		OmitEmpty:               true,
		UsePascalCaseFieldNames: true,
	})
	ts.Decs.Start = nil

	decl := &dst.GenDecl{
		Tok:   token.TYPE,
		Specs: []dst.Spec{ts},
	}

	if schema.Description != "" {
		decl.Decs.Start = docstrings.Emit(schema.Description)
	}

	return decl
}

// EmitType converts a schema to a Go type expression.
// Parameter s is the OpenAPI Schema, and it returns the AST type expression.
func EmitType(s *openapi.Schema) dst.Expr {
	return classes.EmitTypeExpr(s)
}

// toPascalCase delegates identifier conversion to naming.ToPascalCase.
// Parameter s is the input string, and it returns the PascalCase representation.
func toPascalCase(s string) string {
	return naming.ToPascalCase(s)
}
