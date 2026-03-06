package headers

import (
	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
	"strings"
)

// Parse extracts an OpenAPI Header object from a struct field.
func Parse(field *dst.Field) *openapi.Header {
	if field == nil {
		return nil
	}

	header := &openapi.Header{}
	hasContent := false

	if len(field.Decs.Start) > 0 {
		doc := docstrings.Parse(field.Decs.Start)
		lines := strings.Split(doc, "\n")
		var desc []string
		for _, line := range lines {
			lowerLine := strings.ToLower(strings.TrimSpace(line))
			if strings.HasPrefix(lowerLine, "required: true") {
				header.Required = true
				hasContent = true
			} else if lowerLine == "deprecated" {
				header.Deprecated = true
				hasContent = true
			} else {
				desc = append(desc, line)
			}
		}
		header.Description = strings.TrimSpace(strings.Join(desc, "\n"))
		if header.Description != "" {
			hasContent = true
		}
	}

	if ident, ok := field.Type.(*dst.Ident); ok {
		header.Schema = &openapi.Schema{Type: "string"}
		if ident.Name == "int" || ident.Name == "int64" || ident.Name == "int32" {
			header.Schema.Type = "integer"
		} else if ident.Name == "bool" {
			header.Schema.Type = "boolean"
		} else if ident.Name == "float64" || ident.Name == "float32" {
			header.Schema.Type = "number"
		} else if ident.Name != "string" {
			header.Schema = &openapi.Schema{Ref: "#/components/schemas/" + ident.Name}
		}
		hasContent = true
	}

	if !hasContent {
		return nil
	}

	return header
}
