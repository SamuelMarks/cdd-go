package securityschemes

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseSecurityScheme(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("SecuritySchemeApiKey")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Type"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"apiKey"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Description"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"Key Auth"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Name"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"api_key"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("In"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"header"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Scheme"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"bearer"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("BearerFormat"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"JWT"`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("OpenIDConnectURL"),
								Value: &dst.BasicLit{Kind: token.STRING, Value: `"http"`},
							},
							&dst.KeyValueExpr{
								Key: dst.NewIdent("Flows"),
								Value: &dst.CompositeLit{
									Elts: []dst.Expr{
										&dst.KeyValueExpr{
											Key: dst.NewIdent("Implicit"),
											Value: &dst.CompositeLit{
												Elts: []dst.Expr{
													&dst.KeyValueExpr{
														Key:   dst.NewIdent("TokenURL"),
														Value: &dst.BasicLit{Kind: token.STRING, Value: `"t"`},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	name, scheme := Parse(decl)
	if name != "apiKey" {
		t.Errorf("expected apiKey, got %s", name)
	}
	if scheme == nil {
		t.Fatalf("expected scheme")
	}
	if scheme.Type != "apiKey" {
		t.Errorf("expected Type")
	}
	if scheme.Description != "Key Auth" {
		t.Errorf("expected Description")
	}
	if scheme.Name != "api_key" {
		t.Errorf("expected Name")
	}
	if scheme.In != "header" {
		t.Errorf("expected In")
	}
	if scheme.Scheme != "bearer" {
		t.Errorf("expected Scheme")
	}
	if scheme.BearerFormat != "JWT" {
		t.Errorf("expected BearerFormat")
	}
	if scheme.OpenIDConnectURL != "http" {
		t.Errorf("expected URL")
	}
	if scheme.Flows == nil || scheme.Flows.Implicit == nil {
		t.Errorf("expected flows")
	}
}

func TestParseSecuritySchemeNilAndEmpty(t *testing.T) {
	n, s := Parse(nil)
	if n != "" || s != nil {
		t.Errorf("expected nil")
	}

	n, s = Parse(&dst.GenDecl{Tok: token.TYPE})
	if n != "" || s != nil {
		t.Errorf("expected nil for non VAR")
	}

	n, s = Parse(&dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.TypeSpec{}}})
	if n != "" || s != nil {
		t.Errorf("expected nil for non ValueSpec")
	}

	n, s = Parse(&dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("Other")}, Values: []dst.Expr{&dst.CompositeLit{}}}}})
	if n != "" || s != nil {
		t.Errorf("expected nil for non SecurityScheme prefix")
	}
}

func TestParseSecuritySchemeBadFormats(t *testing.T) {
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent("SecuritySchemeEmpty")},
				Values: []dst.Expr{
					&dst.CompositeLit{
						Elts: []dst.Expr{
							dst.NewIdent("bad_kv"),
							&dst.KeyValueExpr{
								Key:   &dst.BasicLit{Kind: token.STRING, Value: `"BadKey"`}, // not an Ident key
								Value: &dst.BasicLit{Kind: token.STRING, Value: `""`},
							},
							&dst.KeyValueExpr{
								Key:   dst.NewIdent("Type"),
								Value: dst.NewIdent("bad_val"), // not a basic lit
							},
						},
					},
				},
			},
		},
	}
	n, s := Parse(decl)
	if n != "" || s != nil {
		t.Errorf("expected nil for bad formats")
	}
}
