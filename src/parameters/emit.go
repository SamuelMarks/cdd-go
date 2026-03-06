package parameters

import (
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats an OpenAPI Parameter object into a struct field AST representation.
func Emit(p openapi.Parameter) *dst.Field {
	var pType dst.Expr = dst.NewIdent("string")
	if p.Schema != nil {
		if p.Schema.Type == "integer" {
			pType = dst.NewIdent("int")
		} else if p.Schema.Type == "boolean" {
			pType = dst.NewIdent("bool")
		} else if p.Schema.Type == "number" {
			pType = dst.NewIdent("float64")
		} else if p.Schema.Ref != "" {
			parts := strings.Split(p.Schema.Ref, "/")
			pType = dst.NewIdent(parts[len(parts)-1])
		}
	} else if p.Ref != "" {
		parts := strings.Split(p.Ref, "/")
		pType = dst.NewIdent(parts[len(parts)-1])
	}

	name := p.Name
	if name == "" && p.Ref != "" {
		parts := strings.Split(p.Ref, "/")
		name = parts[len(parts)-1]
	}

	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  pType,
	}

	var desc []string
	if p.Description != "" {
		desc = append(desc, p.Description)
	}
	if p.In != "" && p.In != "query" && p.In != "path" {
		desc = append(desc, "In: "+p.In)
	}
	if p.Required && p.In != "path" {
		desc = append(desc, "Required: true")
	}
	if p.Deprecated {
		desc = append(desc, "Deprecated")
	}

	if len(desc) > 0 {
		field.Decs.Start = docstrings.Emit(strings.Join(desc, "\n"))
	}

	return field
}
