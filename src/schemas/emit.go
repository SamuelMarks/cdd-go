package schemas

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Schema object into a struct representation.
func Emit(name string, schema *openapi.Schema) *dst.GenDecl {
	if schema == nil || schema.Type != "object" {
		return nil
	}

	st := &dst.StructType{
		Fields: &dst.FieldList{
			List: []*dst.Field{},
		},
	}

	for propName, propSchema := range schema.Properties {
		f := &dst.Field{
			Names: []*dst.Ident{dst.NewIdent(toPascalCase(propName))},
			Type:  EmitType(&propSchema),
			Tag:   &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("`json:\"%s,omitempty\"`", propName)},
		}

		if propSchema.Description != "" {
			f.Decs.Start.Append("// " + propSchema.Description)
		}

		st.Fields.List = append(st.Fields.List, f)
	}

	decl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent(toPascalCase(name)),
				Type: st,
			},
		},
	}

	var comments []string
	if schema.Description != "" {
		comments = append(comments, schema.Description)
	}
	if len(comments) > 0 {
		decl.Decs.Start = docstrings.Emit(strings.Join(comments, "\n"))
	}

	return decl
}

// EmitType converts a schema to a Go type expression
func EmitType(s *openapi.Schema) dst.Expr {
	if s == nil {
		return dst.NewIdent("interface{}")
	}

	if s.Ref != "" {
		parts := strings.Split(s.Ref, "/")
		return dst.NewIdent(toPascalCase(parts[len(parts)-1]))
	}

	switch s.Type {
	case "integer":
		return dst.NewIdent("int")
	case "number":
		return dst.NewIdent("float64")
	case "boolean":
		return dst.NewIdent("bool")
	case "string":
		return dst.NewIdent("string")
	case "array":
		return &dst.ArrayType{Elt: EmitType(s.Items)}
	case "object":
		if s.AdditionalProperties != nil {
			return &dst.MapType{
				Key:   dst.NewIdent("string"),
				Value: EmitType(s.AdditionalProperties),
			}
		}
		return dst.NewIdent("interface{}")
	default:
		return dst.NewIdent("interface{}")
	}
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	var res string
	for _, p := range parts {
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return res
}
