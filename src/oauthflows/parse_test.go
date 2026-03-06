package oauthflows

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseOAuthFlows(t *testing.T) {
	cl := &dst.CompositeLit{
		Elts: []dst.Expr{
			&dst.KeyValueExpr{
				Key: dst.NewIdent("Implicit"),
				Value: &dst.CompositeLit{
					Elts: []dst.Expr{
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("AuthorizationURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"auth"`},
						},
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("TokenURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"token"`},
						},
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("RefreshURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"refresh"`},
						},
						&dst.KeyValueExpr{
							Key: dst.NewIdent("Scopes"),
							Value: &dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   &dst.BasicLit{Kind: token.STRING, Value: `"read"`},
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"Read scope"`},
									},
								},
							},
						},
					},
				},
			},
			&dst.KeyValueExpr{
				Key: dst.NewIdent("Password"),
				Value: &dst.CompositeLit{
					Elts: []dst.Expr{
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("TokenURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"token"`},
						},
					},
				},
			},
			&dst.KeyValueExpr{
				Key: dst.NewIdent("ClientCredentials"),
				Value: &dst.CompositeLit{
					Elts: []dst.Expr{
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("TokenURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"token"`},
						},
					},
				},
			},
			&dst.KeyValueExpr{
				Key: dst.NewIdent("AuthorizationCode"),
				Value: &dst.CompositeLit{
					Elts: []dst.Expr{
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("TokenURL"),
							Value: &dst.BasicLit{Kind: token.STRING, Value: `"token"`},
						},
					},
				},
			},
		},
	}

	flows := Parse(cl)
	if flows == nil {
		t.Fatalf("expected flows")
	}
	if flows.Implicit == nil {
		t.Fatalf("expected implicit")
	}
	if flows.Implicit.AuthorizationURL != "auth" {
		t.Errorf("expected auth URL")
	}
	if flows.Implicit.TokenURL != "token" {
		t.Errorf("expected token URL")
	}
	if flows.Implicit.RefreshURL != "refresh" {
		t.Errorf("expected refresh URL")
	}
	if flows.Implicit.Scopes["read"] != "Read scope" {
		t.Errorf("expected read scope")
	}

	if flows.Password == nil {
		t.Errorf("expected Password")
	}
	if flows.ClientCredentials == nil {
		t.Errorf("expected ClientCredentials")
	}
	if flows.AuthorizationCode == nil {
		t.Errorf("expected AuthorizationCode")
	}
}

func TestParseOAuthFlowsNilAndEmpty(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
	if Parse(&dst.CompositeLit{}) != nil {
		t.Errorf("expected nil for empty lit")
	}
}

func TestParseOAuthFlowsBadFormat(t *testing.T) {
	cl := &dst.CompositeLit{
		Elts: []dst.Expr{
			dst.NewIdent("bad"),
			&dst.KeyValueExpr{
				Key:   &dst.BasicLit{Kind: token.STRING, Value: `"bad_key_type"`},
				Value: &dst.CompositeLit{},
			},
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Implicit"),
				Value: dst.NewIdent("bad_val_type"),
			},
			&dst.KeyValueExpr{
				Key: dst.NewIdent("Implicit"),
				Value: &dst.CompositeLit{
					Elts: []dst.Expr{
						dst.NewIdent("bad_flow_elt"),
						&dst.KeyValueExpr{
							Key:   &dst.BasicLit{Kind: token.STRING, Value: `"bad"`},
							Value: &dst.BasicLit{},
						},
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("AuthorizationURL"),
							Value: dst.NewIdent("bad"),
						},
						&dst.KeyValueExpr{
							Key:   dst.NewIdent("Scopes"),
							Value: dst.NewIdent("bad_scopes_type"),
						},
						&dst.KeyValueExpr{
							Key: dst.NewIdent("Scopes"),
							Value: &dst.CompositeLit{
								Elts: []dst.Expr{
									dst.NewIdent("bad_scope_elt"),
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("bad_scope_key_type"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `""`},
									},
									&dst.KeyValueExpr{
										Key:   &dst.BasicLit{Kind: token.STRING, Value: `""`},
										Value: dst.NewIdent("bad_scope_val_type"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	if Parse(cl) != nil {
		t.Errorf("expected nil when content invalid")
	}
}
