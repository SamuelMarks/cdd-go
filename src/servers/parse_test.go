package servers

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseServers(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("Servers")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("URL"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"http://api"`},
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

	servers, err := Parse(decl)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server")
	}
	if servers[0].URL != "http://api" {
		t.Errorf("expected URL")
	}
	if servers[0].Description != "Desc" {
		t.Errorf("expected description")
	}
}

func TestParseServersNameHandling(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("Servers")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("URL"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"http://api"`},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	servers, err := Parse(decl)
	if err != nil || len(servers) != 1 {
		t.Fatalf("expected servers")
	}
}

func TestParseServersNil(t *testing.T) {
	srvs, err := Parse(nil)
	if err != nil || srvs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseServersNotVar(t *testing.T) {
	srvs, err := Parse(&dst.GenDecl{Tok: token.CONST})
	if err != nil || srvs != nil {
		t.Errorf("expected nil")
	}
}

func TestParseServersEmptyVar(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Other")},
				Values: []dst.Expr{&dst.CompositeLit{}},
			},
		},
	}
	srvs, err := Parse(decl)
	if err != nil || srvs != nil {
		t.Errorf("expected nil")
	}
}
