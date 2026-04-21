package parameters

import (
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Parameter object from a struct field AST representation.
func Parse(field *dst.Field) *openapi.Parameter {
	if field == nil || len(field.Names) == 0 {
		return nil
	}

	p := &openapi.Parameter{
		Name: field.Names[0].Name,
		In:   "query", // default
	}

	if len(field.Decs.Start) > 0 {
		doc := docstrings.Parse(field.Decs.Start)
		lines := strings.Split(doc, "\n")
		var desc []string
		for _, line := range lines {
			lowerLine := strings.ToLower(strings.TrimSpace(line))
			if strings.HasPrefix(lowerLine, "required: true") {
				p.Required = true
			} else if lowerLine == "deprecated" {
				p.Deprecated = true
			} else if strings.HasPrefix(lowerLine, "in:") {
				p.In = strings.TrimSpace(line[3:])
			} else {
				desc = append(desc, line)
			}
		}
		p.Description = strings.TrimSpace(strings.Join(desc, "\n"))
	}

	if strings.Contains(strings.ToLower(p.Name), "id") {
		p.In = "path"
		p.Required = true
	}

	if ident, ok := field.Type.(*dst.Ident); ok {
		p.Schema = &openapi.Schema{Type: "string"}
		if ident.Name == "int" || ident.Name == "int64" || ident.Name == "int32" {
			p.Schema.Type = "integer"
		} else if ident.Name == "bool" {
			p.Schema.Type = "boolean"
		} else if ident.Name == "float64" || ident.Name == "float32" {
			p.Schema.Type = "number"
		} else if ident.Name != "string" {
			p.Schema = &openapi.Schema{Ref: "#/components/schemas/" + ident.Name}
		}
	}

	return p
}
