// Package webhooks provides parsing and emitting of OpenAPI Webhooks maps to and from Go AST declarations.
package webhooks

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Webhooks object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the webhook name and parsed PathItem map.
func Parse(decl *dst.GenDecl) (string, map[string]openapi.PathItem) {
	return astutil.ParsePathItemMapDecl(decl, "Webhook")
}
