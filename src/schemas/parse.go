// Package schemas provides parsing and emitting of OpenAPI Schema objects to and from Go AST declarations.
package schemas

import (
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/classes"
	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Schema object from a type declaration.
// Parameter decl is the AST generic declaration, and it returns the schema name and parsed OpenAPI Schema.
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
							parts := strings.Split(strings.TrimPrefix(tagVal, `json:"`), `"`)
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

// ParseType converts a Go AST type expression to an OpenAPI schema.
// Parameter expr is the AST type expression, and it returns the parsed OpenAPI Schema.
func ParseType(expr dst.Expr) *openapi.Schema {
	s, err := classes.ParseExpr(expr)
	if err != nil || s == nil {
		return &openapi.Schema{}
	}
	return s
}
