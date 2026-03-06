package securityschemes

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/oauthflows"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats an OpenAPI SecurityScheme object into a struct representation.
func Emit(name string, scheme *openapi.SecurityScheme) *dst.GenDecl {
	if scheme == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("Type")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Name")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("In")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Scheme")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("BearerFormat")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("OpenIDConnectURL")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Flows")}, Type: dst.NewIdent("OAuthFlows")},
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

	addField("Type", scheme.Type)
	addField("Description", scheme.Description)
	addField("Name", scheme.Name)
	addField("In", scheme.In)
	addField("Scheme", scheme.Scheme)
	addField("BearerFormat", scheme.BearerFormat)
	addField("OpenIDConnectURL", scheme.OpenIDConnectURL)

	if scheme.Flows != nil {
		flowsLit := oauthflows.Emit(scheme.Flows)
		if flowsLit != nil {
			cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
				Key:   dst.NewIdent("Flows"),
				Value: flowsLit,
			})
		}
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("SecurityScheme" + toPascalCase(name))},
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
