// Package callbacks provides parsing and emitting of OpenAPI Callback objects to and from Go AST declarations.
package callbacks

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Callback object into an AST representation.
func Emit(name string, callback *openapi.Callback) *dst.GenDecl {
	if callback == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: dst.NewIdent("Callback"),
		Elts: []dst.Expr{},
	}

	for k, v := range *callback {
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
				Names:  []*dst.Ident{dst.NewIdent("Callback" + naming.ToPascalCase(name))},
				Values: []dst.Expr{cl},
			},
		},
	}
	return decl
}
