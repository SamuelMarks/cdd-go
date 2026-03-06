package security

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseSecurity(t *testing.T) {
	cl := &dst.CompositeLit{
		Elts: []dst.Expr{
			&dst.CompositeLit{
				Elts: []dst.Expr{
					&dst.KeyValueExpr{
						Key: &dst.BasicLit{Kind: token.STRING, Value: `"api_key"`},
						Value: &dst.CompositeLit{
							Elts: []dst.Expr{},
						},
					},
				},
			},
			&dst.CompositeLit{
				Elts: []dst.Expr{
					&dst.KeyValueExpr{
						Key: &dst.BasicLit{Kind: token.STRING, Value: `"oauth2"`},
						Value: &dst.CompositeLit{
							Elts: []dst.Expr{
								&dst.BasicLit{Kind: token.STRING, Value: `"read"`},
								&dst.BasicLit{Kind: token.STRING, Value: `"write"`},
							},
						},
					},
				},
			},
		},
	}

	sec := Parse(cl)
	if len(sec) != 2 {
		t.Fatalf("expected 2 reqs, got %d", len(sec))
	}

	if scopes, ok := sec[0]["api_key"]; !ok || len(scopes) != 0 {
		t.Errorf("expected empty api_key scopes")
	}

	if scopes, ok := sec[1]["oauth2"]; !ok || len(scopes) != 2 || scopes[0] != "read" || scopes[1] != "write" {
		t.Errorf("expected read/write oauth2 scopes")
	}
}

func TestParseSecurityNilAndEmpty(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
	if Parse(&dst.Ident{}) != nil {
		t.Errorf("expected nil for non composite lit")
	}
	if Parse(&dst.CompositeLit{}) != nil {
		t.Errorf("expected nil for empty lit")
	}
}

func TestParseSecurityBadFormats(t *testing.T) {
	cl := &dst.CompositeLit{
		Elts: []dst.Expr{
			dst.NewIdent("bad_elt"), // not a composite lit map
			&dst.CompositeLit{
				Elts: []dst.Expr{
					dst.NewIdent("bad_kv"), // not a kv expr
					&dst.KeyValueExpr{
						Key:   dst.NewIdent("bad_key"), // not basic lit string
						Value: &dst.CompositeLit{},
					},
					&dst.KeyValueExpr{
						Key:   &dst.BasicLit{Kind: token.STRING, Value: `"key"`},
						Value: dst.NewIdent("bad_val"), // not composite lit slice
					},
					&dst.KeyValueExpr{
						Key: &dst.BasicLit{Kind: token.STRING, Value: `"good"`},
						Value: &dst.CompositeLit{
							Elts: []dst.Expr{
								dst.NewIdent("bad_scope"), // not basic lit string
							},
						},
					},
					&dst.KeyValueExpr{
						Key: &dst.BasicLit{Kind: token.STRING, Value: "unquoted_key"},
						Value: &dst.CompositeLit{
							Elts: []dst.Expr{
								&dst.BasicLit{Kind: token.STRING, Value: "unquoted_val"},
							},
						},
					},
				},
			},
		},
	}

	sec := Parse(cl)
	if len(sec) != 1 {
		t.Fatalf("expected 1 parsed valid element, got %d", len(sec))
	}
	if scopes, ok := sec[0]["good"]; !ok || len(scopes) != 0 {
		t.Errorf("expected good key with 0 scopes")
	}
	if scopes, ok := sec[0]["unquoted_key"]; !ok || len(scopes) != 1 || scopes[0] != "unquoted_val" {
		t.Errorf("expected unquoted values to be mapped")
	}
}
