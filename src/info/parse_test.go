package info

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseInfo(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.CONST,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("Info")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Title"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"Test API"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Version"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"1.0.0"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Description"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"Desc"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Summary"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"Sum"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("TermsOfService"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"TOS"`},
							},
							&dst.KeyValueExpr{
								Key: dst.NewIdent("Contact"),
								Value: &dst.CompositeLit{
									Elts: []dst.Expr{
										&dst.KeyValueExpr{Key: dst.NewIdent("Name"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"Name"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("URL"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"URL"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("Email"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"Email"`}},
									},
								},
							},
							&dst.KeyValueExpr{
								Key: dst.NewIdent("License"),
								Value: &dst.CompositeLit{
									Elts: []dst.Expr{
										&dst.KeyValueExpr{Key: dst.NewIdent("Name"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"LName"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("URL"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"LURL"`}},
										&dst.KeyValueExpr{Key: dst.NewIdent("Identifier"), Value: &dst.BasicLit{Kind: token.STRING, Value: `"LId"`}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	info, err := Parse(decl)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if info == nil {
		t.Fatalf("expected info")
	}
	if info.Title != "Test API" {
		t.Errorf("expected Title")
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected Version")
	}
	if info.Contact == nil || info.Contact.Name != "Name" {
		t.Errorf("expected Contact")
	}
	if info.License == nil || info.License.Name != "LName" {
		t.Errorf("expected License")
	}
}

func TestParseInfoNil(t *testing.T) {
	info, err := Parse(nil)
	if err != nil || info != nil {
		t.Errorf("expected nil")
	}
}

func TestParseInfoNotConst(t *testing.T) {
	info, err := Parse(&dst.GenDecl{Tok: token.VAR})
	if err != nil || info != nil {
		t.Errorf("expected nil")
	}
}

func TestParseInfoEmpty(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.CONST,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Other")},
				Values: []dst.Expr{&dst.CompositeLit{}},
			},
		},
	}
	info, err := Parse(decl)
	if err != nil || info != nil {
		t.Errorf("expected nil")
	}
}

func TestParseInfoUnquoteError(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.CONST,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("Info")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							dst.NewIdent("bad"),
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Title"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `bad_quote`},
							},
						},
					},
				},
			},
		},
	}
	info, err := Parse(decl)
	if err != nil || info == nil {
		t.Errorf("expected info object even on bad quote")
	}
}
