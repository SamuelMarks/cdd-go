package requestbodies

import (
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI RequestBody object into a struct field AST representation.
func Emit(rb *openapi.RequestBody) *dst.Field {
	if rb == nil {
		return nil
	}

	var rbType dst.Expr = dst.NewIdent("any")

	if rb.Content != nil {
		if mt, ok := rb.Content["application/json"]; ok && mt.Schema != nil {
			if mt.Schema.Ref != "" {
				parts := strings.Split(mt.Schema.Ref, "/")
				rbType = dst.NewIdent(parts[len(parts)-1])
			} else if mt.Schema.Type == "array" {
				if mt.Schema.Items != nil && mt.Schema.Items.Ref != "" {
					parts := strings.Split(mt.Schema.Items.Ref, "/")
					rbType = &dst.ArrayType{Elt: dst.NewIdent(parts[len(parts)-1])}
				}
			} else if mt.Schema.Type == "object" {
				rbType = &dst.MapType{
					Key:   dst.NewIdent("string"),
					Value: dst.NewIdent("any"),
				}
			}
		}
	}

	field := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent("body")},
		Type:  rbType,
	}

	var desc []string
	if rb.Description != "" {
		desc = append(desc, rb.Description)
	}
	if rb.Required {
		desc = append(desc, "Required: true")
	}

	if len(desc) > 0 {
		field.Decs.Start = docstrings.Emit(strings.Join(desc, "\n"))
	}

	return field
}
