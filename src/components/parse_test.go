package components

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseComponents(t *testing.T) {
	file := &dst.File{
		Name: dst.NewIdent("components"),
		Decls: []dst.Decl{
			&dst.GenDecl{
				Tok: token.VAR,
				Specs: []dst.Spec{
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("SecuritySchemeBasic")},
						Values: []dst.Expr{
							&dst.CompositeLit{
								Elts: []dst.Expr{
									&dst.KeyValueExpr{
										Key:   dst.NewIdent("Type"),
										Value: &dst.BasicLit{Kind: token.STRING, Value: `"http"`},
									},
								},
							},
						},
					},
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("ParamLimit")},
						Type:  dst.NewIdent("int"),
					},
					&dst.ValueSpec{
						Names: []*dst.Ident{dst.NewIdent("HeaderRateLimit")},
						Type:  dst.NewIdent("int"),
					},
				},
			},
			&dst.GenDecl{
				Tok: token.TYPE,
				Specs: []dst.Spec{
					&dst.TypeSpec{
						Name: dst.NewIdent("RequestBodyUserBody"),
						Type: dst.NewIdent("User"),
					},
					&dst.TypeSpec{
						Name: dst.NewIdent("ResponseUserResp"),
						Type: &dst.StarExpr{X: dst.NewIdent("User")},
					},
					&dst.TypeSpec{
						Name: dst.NewIdent("User"),
						Type: &dst.StructType{
							Fields: &dst.FieldList{
								List: []*dst.Field{
									{Names: []*dst.Ident{dst.NewIdent("Id")}, Type: dst.NewIdent("string")},
								},
							},
						},
					},
				},
			},
		},
	}

	comp := Parse(file)
	if comp == nil {
		t.Fatalf("expected components")
	}

	if s, ok := comp.SecuritySchemes["basic"]; !ok || s.Type != "http" {
		t.Errorf("missing scheme")
	}
	if p, ok := comp.Parameters["limit"]; !ok || p.Schema.Type != "integer" {
		t.Errorf("missing param")
	}
	if h, ok := comp.Headers["rateLimit"]; !ok || h.Schema.Type != "integer" {
		t.Errorf("missing header")
	}
	if rb, ok := comp.RequestBodies["userBody"]; !ok || rb.Content["application/json"].Schema.Ref != "#/components/schemas/User" {
		t.Errorf("missing req body")
	}
	if r, ok := comp.Responses["userResp"]; !ok || r.Content["application/json"].Schema.Ref != "#/components/schemas/User" {
		t.Errorf("missing resp")
	}
	if s, ok := comp.Schemas["User"]; !ok || s.Type != "object" {
		t.Errorf("missing schema")
	}
}

func TestParseComponentsComplexResponse(t *testing.T) {
	file := &dst.File{
		Decls: []dst.Decl{
			&dst.GenDecl{
				Tok: token.TYPE,
				Specs: []dst.Spec{
					&dst.TypeSpec{
						Name: dst.NewIdent("ResponseComplex"),
						Type: &dst.StructType{
							Fields: &dst.FieldList{
								List: []*dst.Field{
									{Type: &dst.StarExpr{X: dst.NewIdent("User")}},
									{Names: []*dst.Ident{dst.NewIdent("X-RateLimit")}, Type: dst.NewIdent("int")},
								},
							},
						},
					},
				},
			},
		},
	}

	comp := Parse(file)
	if r, ok := comp.Responses["complex"]; !ok {
		t.Errorf("missing complex resp")
	} else {
		if r.Content["application/json"].Schema.Ref != "#/components/schemas/User" {
			t.Errorf("missing User ref")
		}
		if len(r.Headers) != 1 {
			t.Errorf("missing header")
		}
	}
}

func TestParseComponentsNilAndEmpty(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
	if Parse(&dst.File{}) != nil {
		t.Errorf("expected nil for empty file")
	}
}
