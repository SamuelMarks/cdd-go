package requestbodies

import (
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI RequestBody object from a struct field AST representation.
func Parse(field *dst.Field) *openapi.RequestBody {
	if field == nil {
		return nil
	}

	rb := &openapi.RequestBody{
		Content: map[string]openapi.MediaType{
			"application/json": {
				Schema: &openapi.Schema{},
			},
		},
	}

	if len(field.Decs.Start) > 0 {
		doc := docstrings.Parse(field.Decs.Start)
		lines := strings.Split(doc, "\n")
		var desc []string
		for _, line := range lines {
			lowerLine := strings.ToLower(strings.TrimSpace(line))
			if strings.HasPrefix(lowerLine, "required: true") {
				rb.Required = true
			} else {
				desc = append(desc, line)
			}
		}
		rb.Description = strings.TrimSpace(strings.Join(desc, "\n"))
	}

	if ident, ok := field.Type.(*dst.Ident); ok {
		if ident.Name == "any" || ident.Name == "interface{}" {
			rb.Content["application/json"].Schema.Type = "object"
		} else if ident.Name != "string" && ident.Name != "int" && ident.Name != "bool" && ident.Name != "float64" {
			rb.Content["application/json"].Schema.Ref = "#/components/schemas/" + ident.Name
		} else {
			rb.Content["application/json"].Schema.Type = ident.Name
		}
	} else if arrayType, ok := field.Type.(*dst.ArrayType); ok {
		rb.Content["application/json"].Schema.Type = "array"
		if itemIdent, ok := arrayType.Elt.(*dst.Ident); ok {
			rb.Content["application/json"].Schema.Items = &openapi.Schema{Ref: "#/components/schemas/" + itemIdent.Name}
		}
	} else if starType, ok := field.Type.(*dst.StarExpr); ok {
		if ident, ok := starType.X.(*dst.Ident); ok {
			rb.Content["application/json"].Schema.Ref = "#/components/schemas/" + ident.Name
		}
	} else {
		rb.Content["application/json"].Schema.Type = "object"
	}

	return rb
}
