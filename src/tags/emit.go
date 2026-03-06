package tags

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats an OpenAPI Tags slice into an AST representation.
func Emit(tags []openapi.Tag) *dst.GenDecl {
	if len(tags) == 0 {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.ArrayType{
			Elt: &dst.StructType{
				Fields: &dst.FieldList{
					List: []*dst.Field{
						{Names: []*dst.Ident{dst.NewIdent("Name")}, Type: dst.NewIdent("string")},
						{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
					},
				},
			},
		},
		Elts: []dst.Expr{},
	}

	for _, tag := range tags {
		tagLit := &dst.CompositeLit{
			Elts: []dst.Expr{
				&dst.KeyValueExpr{
					Key:   dst.NewIdent("Name"),
					Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", tag.Name)},
				},
			},
		}

		if tag.Description != "" {
			tagLit.Elts = append(tagLit.Elts, &dst.KeyValueExpr{
				Key:   dst.NewIdent("Description"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", tag.Description)},
			})
		}

		cl.Elts = append(cl.Elts, tagLit)
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Tags")},
				Values: []dst.Expr{cl},
			},
		},
	}

	return decl
}
