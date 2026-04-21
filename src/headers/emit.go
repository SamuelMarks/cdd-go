package headers

import (
	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"strings"
)

// Emit formats an OpenAPI Header object into a struct field.
func Emit(name string, header *openapi.Header) *dst.Field {
	if header == nil {
		return nil
	}

	var fType dst.Expr = dst.NewIdent("string")
	if header.Schema != nil {
		if header.Schema.Type == "integer" {
			fType = dst.NewIdent("int")
		} else if header.Schema.Type == "boolean" {
			fType = dst.NewIdent("bool")
		} else if header.Schema.Type == "number" {
			fType = dst.NewIdent("float64")
		} else if header.Schema.Ref != "" {
			parts := strings.Split(header.Schema.Ref, "/")
			fType = dst.NewIdent(parts[len(parts)-1])
		}
	}

	f := &dst.Field{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  fType,
	}

	var desc []string
	if header.Description != "" {
		desc = append(desc, header.Description)
	}
	if header.Required {
		desc = append(desc, "Required: true")
	}
	if header.Deprecated {
		desc = append(desc, "Deprecated")
	}

	if len(desc) > 0 {
		f.Decs.Start = docstrings.Emit(strings.Join(desc, "\n"))
	}

	return f
}
