package commands

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseCommand(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// Method: GET",
					"// Path: /users/{id}",
				},
			},
		},
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("GetUserCmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X: &dst.CompositeLit{
							Elts: []dst.Expr{
								&dst.KeyValueExpr{
									Key:   dst.NewIdent("Use"),
									Value: &dst.BasicLit{Kind: token.STRING, Value: `"getuser"`},
								},
								&dst.KeyValueExpr{
									Key:   dst.NewIdent("Short"),
									Value: &dst.BasicLit{Kind: token.STRING, Value: `"Get user"`},
								},
								&dst.KeyValueExpr{
									Key:   dst.NewIdent("Long"),
									Value: &dst.BasicLit{Kind: token.STRING, Value: `"Gets a user by ID"`},
								},
							},
						},
					},
				},
			},
		},
	}

	method, path, op := Parse(decl)
	if method != "get" {
		t.Errorf("expected get, got %s", method)
	}
	if path != "/users/{id}" {
		t.Errorf("expected /users/{id}, got %s", path)
	}
	if op == nil {
		t.Fatalf("expected op")
	}
	if op.OperationID != "getUser" {
		t.Errorf("expected getUser, got %s", op.OperationID)
	}
	if op.Summary != "Get user" {
		t.Errorf("expected summary")
	}
	if op.Description != "Gets a user by ID" {
		t.Errorf("expected description")
	}
}

func TestParseCommandNilAndEmpty(t *testing.T) {
	m, p, op := Parse(nil)
	if m != "" || p != "" || op != nil {
		t.Errorf("expected nil")
	}

	m, p, op = Parse(&dst.GenDecl{Tok: token.TYPE})
	if m != "" || p != "" || op != nil {
		t.Errorf("expected nil for non VAR")
	}

	m, p, op = Parse(&dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.TypeSpec{}}})
	if m != "" || p != "" || op != nil {
		t.Errorf("expected nil for non ValueSpec")
	}

	m, p, op = Parse(&dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("Other")}, Values: []dst.Expr{&dst.CompositeLit{}}}}})
	if m != "" || p != "" || op != nil {
		t.Errorf("expected nil for non Cmd prefix")
	}
}

func TestParseCommandBadFormats(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("CmdEmptyCmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X: &dst.CompositeLit{
							Elts: []dst.Expr{
								dst.NewIdent("bad_kv"),
								&dst.KeyValueExpr{
									Key:   &dst.BasicLit{Kind: token.STRING, Value: `"BadKey"`}, // not an Ident key
									Value: &dst.BasicLit{Kind: token.STRING, Value: `""`},
								},
								&dst.KeyValueExpr{
									Key:   dst.NewIdent("Use"),
									Value: dst.NewIdent("bad_val"), // not a basic lit
								},
							},
						},
					},
				},
			},
		},
	}
	m, p, op := Parse(decl)
	if m != "" || p != "" || op != nil {
		t.Errorf("expected nil for bad formats")
	}
}
