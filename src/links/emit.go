package links

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Link object into an AST representation.
func Emit(name string, link *openapi.Link) *dst.GenDecl {
	if link == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("OperationRef")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("OperationID")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
				},
			},
		},
		Elts: []dst.Expr{},
	}

	addField := func(key, value string) {
		if value != "" {
			cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
				Key:   dst.NewIdent(key),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", value)},
			})
		}
	}

	addField("OperationRef", link.OperationRef)
	addField("OperationID", link.OperationID)
	addField("Description", link.Description)

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Link" + toPascalCase(name))},
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
