package responses

import (
	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/headers"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI Responses object from a Go return type list.
func Parse(results *dst.FieldList) openapi.Responses {
	if results == nil {
		return nil
	}

	resps := openapi.Responses{}
	hasConcreteType := false
	var parsedHeaders map[string]openapi.Header

	for _, field := range results.List {
		if ident, ok := field.Type.(*dst.Ident); ok && ident.Name == "error" {
			continue
		}

		var typeName string
		if star, ok := field.Type.(*dst.StarExpr); ok {
			if ident, ok := star.X.(*dst.Ident); ok {
				typeName = ident.Name
			}
		} else if ident, ok := field.Type.(*dst.Ident); ok {
			typeName = ident.Name
		}

		if typeName != "" && typeName != "Response" && typeName != "Context" && typeName != "string" && typeName != "int" && typeName != "bool" && typeName != "float64" {
			resps["200"] = openapi.Response{
				Description: "Success",
				Content: map[string]openapi.MediaType{
					"application/json": {
						Schema: &openapi.Schema{
							Ref: "#/components/schemas/" + typeName,
						},
					},
				},
			}
			hasConcreteType = true
			continue
		}

		// Try parsing as header
		h := headers.Parse(field)
		if h != nil && len(field.Names) > 0 {
			if parsedHeaders == nil {
				parsedHeaders = make(map[string]openapi.Header)
			}
			parsedHeaders[field.Names[0].Name] = *h
			hasConcreteType = true
		}
	}

	if hasConcreteType {
		if r, ok := resps["200"]; ok {
			if len(parsedHeaders) > 0 {
				r.Headers = parsedHeaders
				resps["200"] = r
			}
		} else if len(parsedHeaders) > 0 {
			resps["200"] = openapi.Response{
				Description: "Success",
				Headers:     parsedHeaders,
			}
		}
	} else {
		resps["default"] = openapi.Response{
			Description: "Default response",
		}
	}
	return resps
}
