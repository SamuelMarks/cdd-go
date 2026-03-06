package tags

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseTags(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("Tags")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("Name"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"users"`},
									},
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("Description"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"Desc"`},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tags, err := Parse(decl)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag")
	}
	if tags[0].Name != "users" {
		t.Errorf("expected Name")
	}
	if tags[0].Description != "Desc" {
		t.Errorf("expected description")
	}
}

func TestParseTagsNil(t *testing.T) {
	tgs, err := Parse(nil)
	if err != nil || tgs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseTagsNotVar(t *testing.T) {
	tgs, err := Parse(&dst.GenDecl{Tok: token.CONST})
	if err != nil || tgs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseTagsEmptyVar(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Other")},
				Values: []dst.Expr{&dst.CompositeLit{}},
			},
		},
	}
	tgs, err := Parse(decl)
	if err != nil || tgs != nil {
		t.Errorf("expected nil")
	}
}
