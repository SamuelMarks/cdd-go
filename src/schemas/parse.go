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

// Parse extracts an OpenAPI Schema object from a Go AST type declaration.
// Parameter decl is the AST generic declaration, and it returns the schema name, parsed OpenAPI Schema, and error.
func Parse(decl *dst.GenDecl) (string, *openapi.Schema, error) {
	if decl == nil || decl.Tok != token.TYPE {
		return "", nil, nil
	}

	for _, spec := range decl.Specs {
		ts, ok := spec.(*dst.TypeSpec)
		if !ok {
			continue
		}
		name := ts.Name.Name

		schema, err := classes.ParseType(ts)
		if err != nil {
			return "", nil, err
		}

		if len(decl.Decs.Start) > 0 {
			schema.Description = strings.TrimSpace(docstrings.Parse(decl.Decs.Start))
		}

		if st, ok := ts.Type.(*dst.StructType); ok && st.Fields != nil {
			for _, f := range st.Fields.List {
				if len(f.Names) == 0 {
					continue
				}
				rawName := f.Names[0].Name
				lowerName := strings.ToLower(rawName[:1]) + rawName[1:]
				hasJSONTag := f.Tag != nil && strings.Contains(f.Tag.Value, "json:")
				if !hasJSONTag && rawName != lowerName {
					if prop, exists := schema.Properties[rawName]; exists {
						delete(schema.Properties, rawName)
						schema.Properties[lowerName] = prop
					}
				}
			}
		}

		return name, schema, nil
	}

	return "", nil, nil
}

// ParseType converts a Go AST type expression to an OpenAPI schema.
// Parameter expr is the AST type expression, and it returns the parsed OpenAPI Schema and error.
func ParseType(expr dst.Expr) (*openapi.Schema, error) {
	return classes.ParseExpr(expr)
}
