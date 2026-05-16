package webhooks

import (
	"fmt"
	"go/token"

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

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
