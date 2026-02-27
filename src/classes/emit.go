// Package classes provides parsing and emitting of Go structs from/to OpenAPI schemas.
package classes

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// EmitType converts an OpenAPI Schema into a dst.TypeSpec.
func EmitType(name string, schema *openapi.Schema) (*dst.TypeSpec, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	ts := &dst.TypeSpec{
		Name: dst.NewIdent(name),
	}

	if schema.Type == "object" || len(schema.Properties) > 0 {
		st := &dst.StructType{
			Fields: &dst.FieldList{},
		}
		for propName, propSchema := range schema.Properties {
			pSchema := propSchema // create local copy
			f := &dst.Field{
				Names: []*dst.Ident{dst.NewIdent(exportedName(propName))},
				Type:  EmitTypeExpr(&pSchema),
				Tag: &dst.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("`json:\"%s\"`", propName),
				},
			}
			if pSchema.Description != "" {
				f.Decs.Start.Append(fmt.Sprintf("// %s", pSchema.Description))
			}
			st.Fields.List = append(st.Fields.List, f)
		}
		ts.Type = st
	} else if schema.Type == "array" {
		at := &dst.ArrayType{
			Elt: EmitTypeExpr(schema.Items),
		}
		ts.Type = at
	} else {
		ts.Type = EmitTypeExpr(schema)
	}

	if schema.Description != "" {
		ts.Decs.Start.Append(fmt.Sprintf("// %s", schema.Description))
	}

	return ts, nil
}

// EmitTypeExpr returns a dst expression for the schema's type.
func EmitTypeExpr(schema *openapi.Schema) dst.Expr {
	if schema == nil {
		return dst.NewIdent("interface{}")
	}
	switch schema.Type {
	case "string":
		return dst.NewIdent("string")
	case "integer":
		return dst.NewIdent("int")
	case "number":
		return dst.NewIdent("float64")
	case "boolean":
		return dst.NewIdent("bool")
	case "array":
		return &dst.ArrayType{
			Elt: EmitTypeExpr(schema.Items),
		}
	default:
		if schema.Ref != "" {
			parts := strings.Split(schema.Ref, "/")
			refName := parts[len(parts)-1]
			return dst.NewIdent(refName)
		}
		return dst.NewIdent("interface{}")
	}
}

// exportedName converts a json field name to an exported Go field name.
func exportedName(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + name[1:]
}
