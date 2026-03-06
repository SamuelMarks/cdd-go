package externaldocs

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseExternalDocs(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("ExternalDocs")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Description"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"Desc"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("URL"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"http://api"`},
							},
						},
					},
				},
			},
		},
	}

	docs, err := Parse(decl)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if docs == nil {
		t.Fatalf("expected docs")
	}
	if docs.Description != "Desc" {
		t.Errorf("expected description")
	}
	if docs.URL != "http://api" {
		t.Errorf("expected URL")
	}
}

func TestParseExternalDocsNil(t *testing.T) {
	docs, err := Parse(nil)
	if err != nil || docs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseExternalDocsNotVar(t *testing.T) {
	docs, err := Parse(&dst.GenDecl{Tok: token.CONST})
	if err != nil || docs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseExternalDocsEmptyVar(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Other")},
				Values: []dst.Expr{&dst.CompositeLit{}},
			},
		},
	}
	docs, err := Parse(decl)
	if err != nil || docs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseExternalDocsEmpty(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("ExternalDocs")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							dst.NewIdent("bad"),
						},
					},
				},
			},
		},
	}
	docs, err := Parse(decl)
	if err != nil || docs != nil {
		t.Errorf("expected nil")
	}
}
