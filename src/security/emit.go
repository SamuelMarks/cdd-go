package security

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats OpenAPI SecurityRequirement objects into a structured AST slice.
func Emit(security []openapi.SecurityRequirement) *dst.CompositeLit {
	if len(security) == 0 {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.ArrayType{
			Elt: &dst.MapType{
				Key: dst.NewIdent("string"),
				Value: &dst.ArrayType{
					Elt: dst.NewIdent("string"),
				},
			},
		},
		Elts: []dst.Expr{},
	}

	for _, secReq := range security {
		mapLit := &dst.CompositeLit{
			Type: &dst.MapType{
				Key: dst.NewIdent("string"),
				Value: &dst.ArrayType{
					Elt: dst.NewIdent("string"),
				},
			},
			Elts: []dst.Expr{},
		}

		for key, scopes := range secReq {
			scopesLit := &dst.CompositeLit{
				Type: &dst.ArrayType{
					Elt: dst.NewIdent("string"),
				},
				Elts: []dst.Expr{},
			}
			for _, scope := range scopes {
				scopesLit.Elts = append(scopesLit.Elts, &dst.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", scope),
				})
			}

			mapLit.Elts = append(mapLit.Elts, &dst.KeyValueExpr{
				Key: &dst.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", key),
				},
				Value: scopesLit,
			})
		}
		cl.Elts = append(cl.Elts, mapLit)
	}

	return cl
}
