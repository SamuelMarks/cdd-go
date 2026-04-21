package servers

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Servers slice into an AST representation.
func Emit(servers []openapi.Server) *dst.GenDecl {
	if len(servers) == 0 {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.ArrayType{
			Elt: &dst.StructType{
				Fields: &dst.FieldList{
					List: []*dst.Field{
						{Names: []*dst.Ident{dst.NewIdent("URL")}, Type: dst.NewIdent("string")},
						{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
					},
				},
			},
		},
		Elts: []dst.Expr{},
	}

	for _, srv := range servers {
		serverLit := &dst.CompositeLit{
			Elts: []dst.Expr{
				&dst.KeyValueExpr{
					Key:   dst.NewIdent("URL"),
					Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", srv.URL)},
				},
			},
		}

		if srv.Description != "" {
			serverLit.Elts = append(serverLit.Elts, &dst.KeyValueExpr{
				Key:   dst.NewIdent("Description"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", srv.Description)},
			})
		}

		cl.Elts = append(cl.Elts, serverLit)
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Servers")},
				Values: []dst.Expr{cl},
			},
		},
	}

	return decl
}
