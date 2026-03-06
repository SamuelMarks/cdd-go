package externaldocs

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats an OpenAPI ExternalDocs object into an AST representation.
func Emit(docs *openapi.ExternalDocs) (*dst.GenDecl, error) {
	if docs == nil {
		return nil, nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("URL")}, Type: dst.NewIdent("string")},
				},
			},
		},
		Elts: []dst.Expr{},
	}

	if docs.Description != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Description"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", docs.Description)},
		})
	}

	if docs.URL != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("URL"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", docs.URL)},
		})
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("ExternalDocs")},
				Values: []dst.Expr{cl},
			},
		},
	}

	return decl, nil
}
