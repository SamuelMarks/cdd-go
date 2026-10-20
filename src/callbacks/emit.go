// Package callbacks provides parsing and emitting of OpenAPI Callback objects to and from Go AST declarations.
package callbacks

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Callback object into an AST representation.
// Parameter name is the callback identifier, callback is the OpenAPI Callback map, and it returns the AST variable declaration.
func Emit(name string, callback *openapi.Callback) *dst.GenDecl {
	if callback == nil {
		return nil
	}
	return astutil.EmitPathItemMapDecl("Callback", "Callback", name, map[string]openapi.PathItem(*callback))
}
