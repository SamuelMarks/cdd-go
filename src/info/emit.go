package info

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Emit formats an OpenAPI Info object into an AST representation.
func Emit(info openapi.Info) *dst.GenDecl {
	cl := &dst.CompositeLit{
		Type: &dst.StructType{
			Fields: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("Title")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Version")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Description")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("TermsOfService")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("Summary")}, Type: dst.NewIdent("string")},
					{
						Names: []*dst.Ident{dst.NewIdent("Contact")},
						Type: &dst.StructType{
							Fields: &dst.FieldList{
								List: []*dst.Field{
									{Names: []*dst.Ident{dst.NewIdent("Name")}, Type: dst.NewIdent("string")},
									{Names: []*dst.Ident{dst.NewIdent("URL")}, Type: dst.NewIdent("string")},
									{Names: []*dst.Ident{dst.NewIdent("Email")}, Type: dst.NewIdent("string")},
								},
							},
						},
					},
					{
						Names: []*dst.Ident{dst.NewIdent("License")},
						Type: &dst.StructType{
							Fields: &dst.FieldList{
								List: []*dst.Field{
									{Names: []*dst.Ident{dst.NewIdent("Name")}, Type: dst.NewIdent("string")},
									{Names: []*dst.Ident{dst.NewIdent("URL")}, Type: dst.NewIdent("string")},
									{Names: []*dst.Ident{dst.NewIdent("Identifier")}, Type: dst.NewIdent("string")},
								},
							},
						},
					},
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

	addField("Title", info.Title)
	addField("Version", info.Version)
	addField("Description", info.Description)
	addField("Summary", info.Summary)
	addField("TermsOfService", info.TermsOfService)

	if info.Contact != nil {
		contactLit := &dst.CompositeLit{Elts: []dst.Expr{}}
		if info.Contact.Name != "" {
			contactLit.Elts = append(contactLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("Name"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Contact.Name)}})
		}
		if info.Contact.URL != "" {
			contactLit.Elts = append(contactLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("URL"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Contact.URL)}})
		}
		if info.Contact.Email != "" {
			contactLit.Elts = append(contactLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("Email"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.Contact.Email)}})
		}
		if len(contactLit.Elts) > 0 {
			cl.Elts = append(cl.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("Contact"), Value: contactLit})
		}
	}

	if info.License != nil {
		licenseLit := &dst.CompositeLit{Elts: []dst.Expr{}}
		if info.License.Name != "" {
			licenseLit.Elts = append(licenseLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("Name"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.License.Name)}})
		}
		if info.License.URL != "" {
			licenseLit.Elts = append(licenseLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("URL"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.License.URL)}})
		}
		if info.License.Identifier != "" {
			licenseLit.Elts = append(licenseLit.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("Identifier"), Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", info.License.Identifier)}})
		}
		if len(licenseLit.Elts) > 0 {
			cl.Elts = append(cl.Elts, &dst.KeyValueExpr{Key: dst.NewIdent("License"), Value: licenseLit})
		}
	}

	if len(cl.Elts) == 0 {
		return nil
	}

	decl := &dst.GenDecl{
		Tok: token.CONST,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Info")},
				Values: []dst.Expr{cl},
			},
		},
	}

	return decl
}
