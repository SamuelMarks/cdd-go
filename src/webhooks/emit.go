// Package webhooks provides parsing and emitting of OpenAPI Webhooks maps to and from Go AST declarations.
package webhooks

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Webhooks object into an AST representation.
// Parameter name is the webhook identifier, webhooks is the PathItem map, and it returns the AST variable declaration.
func Emit(name string, webhooks map[string]openapi.PathItem) *dst.GenDecl {
	return astutil.EmitPathItemMapDecl("Webhooks", "Webhook", name, webhooks)
}
