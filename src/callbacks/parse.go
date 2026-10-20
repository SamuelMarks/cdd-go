// Package callbacks provides parsing and emitting of OpenAPI Callback objects to and from Go AST declarations.
package callbacks

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Callback object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the callback name and parsed Callback object.
func Parse(decl *dst.GenDecl) (string, *openapi.Callback) {
	name, items := astutil.ParsePathItemMapDecl(decl, "Callback")
	if items == nil {
		return "", nil
	}
	cb := openapi.Callback(items)
	return name, &cb
}
