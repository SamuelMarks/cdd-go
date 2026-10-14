// Package webhooks provides parsing and emitting of OpenAPI Webhooks maps to and from Go AST declarations.
package webhooks

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Webhooks object into an AST representation.
func Emit(name string, webhooks map[string]openapi.PathItem) *dst.GenDecl {
	if webhooks == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: dst.NewIdent("Webhooks"),
		Elts: []dst.Expr{},
	}

	for k, v := range webhooks {
		valLit := &dst.CompositeLit{
			Type: dst.NewIdent("PathItem"),
			Elts: []dst.Expr{
				&dst.KeyValueExpr{
					Key:   dst.NewIdent("Summary"),
					Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", v.Summary)},
				},
			},
		}
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", k)},
			Value: valLit,
		})
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Webhook" + toPascalCase(name))},
				Values: []dst.Expr{cl},
			},
		},
	}
	return decl
}

// toPascalCase converts an identifier string to PascalCase.
// Parameter s is the input string, and it returns the PascalCase representation.
func toPascalCase(s string) string {
	return naming.ToPascalCase(s)
}
