package schemas

import (
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI Schema object from a type declaration
func Parse(decl *dst.GenDecl) (string, *openapi.Schema) {
	if decl == nil || decl.Tok != token.TYPE {
		return "", nil
	}

	for _, spec := range decl.Specs {
		if ts, ok := spec.(*dst.TypeSpec); ok {
			name := ts.Name.Name

			schema := &openapi.Schema{
				Type:       "object",
				Properties: make(map[string]openapi.Schema),
			}

			if len(decl.Decs.Start) > 0 {
				schema.Description = strings.TrimSpace(docstrings.Parse(decl.Decs.Start))
			}

			if st, ok := ts.Type.(*dst.StructType); ok && st.Fields != nil {
				for _, f := range st.Fields.List {
					if len(f.Names) == 0 {
						continue
					}

					propName := f.Names[0].Name
					propName = strings.ToLower(propName[:1]) + propName[1:]

					if f.Tag != nil {
						tagVal := strings.Trim(f.Tag.Value, "`")
						if strings.HasPrefix(tagVal, "json:") {
							parts := strings.Split(strings.TrimPrefix(tagVal, "json:\""), "\"")
							if len(parts) > 0 {
								jsonName := strings.Split(parts[0], ",")[0]
								if jsonName != "" && jsonName != "-" {
									propName = jsonName
								}
							}
						}
					}

					propSchema := ParseType(f.Type)
					if len(f.Decs.Start) > 0 {
						propSchema.Description = strings.TrimSpace(docstrings.Parse(f.Decs.Start))
					}

					schema.Properties[propName] = *propSchema
				}

				return name, schema
			} else if mt, ok := ts.Type.(*dst.MapType); ok {
				schema.Type = "object"
				schema.Properties = nil
				schema.AdditionalProperties = ParseType(mt.Value)
				return name, schema
			}
		}
	}

	return "", nil
}

// ParseType converts a Go AST type expression to an OpenAPI schema
func ParseType(expr dst.Expr) *openapi.Schema {
	if expr == nil {
		return &openapi.Schema{}
	}

	switch e := expr.(type) {
	case *dst.Ident:
		switch e.Name {
		case "int", "int32", "int64":
			return &openapi.Schema{Type: "integer"}
		case "float32", "float64":
			return &openapi.Schema{Type: "number"}
		case "bool":
			return &openapi.Schema{Type: "boolean"}
		case "string":
			return &openapi.Schema{Type: "string"}
		case "any":
			return &openapi.Schema{}
		default:
			return &openapi.Schema{Ref: "#/components/schemas/" + e.Name}
		}
	case *dst.StarExpr:
		return ParseType(e.X)
	case *dst.ArrayType:
		return &openapi.Schema{Type: "array", Items: ParseType(e.Elt)}
	case *dst.MapType:
		return &openapi.Schema{Type: "object", AdditionalProperties: ParseType(e.Value)}
	case *dst.SelectorExpr:
		if x, ok := e.X.(*dst.Ident); ok && x.Name == "time" && e.Sel.Name == "Time" {
			return &openapi.Schema{Type: "string", Format: "date-time"}
		}
	}

	return &openapi.Schema{}
}
