package astutil

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestFindValueCompositeLit(t *testing.T) {
	if n, cl := FindValueCompositeLit(nil, "Info", ""); n != "" || cl != nil {
		t.Error("expected empty for nil decl")
	}
	if n, cl := FindValueCompositeLit(&dst.GenDecl{Tok: token.TYPE}, "Info", ""); n != "" || cl != nil {
		t.Error("expected empty for TYPE decl")
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.TypeSpec{Name: dst.NewIdent("Ignored")},
			&dst.ValueSpec{},
			&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("Other")}, Values: []dst.Expr{&dst.CompositeLit{}}},
			&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("LinkNonLit")}, Values: []dst.Expr{dst.NewIdent("x")}},
			&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("LinkUser")}, Values: []dst.Expr{&dst.CompositeLit{}}},
		},
	}

	if n, cl := FindValueCompositeLit(decl, "Missing", ""); n != "" || cl != nil {
		t.Error("expected nil for missing exact match")
	}
	if n, cl := FindValueCompositeLit(decl, "Other", ""); n != "Other" || cl == nil {
		t.Errorf("expected Other match, got %q, %v", n, cl)
	}
	if n, cl := FindValueCompositeLit(decl, "", "Link"); n != "LinkUser" || cl == nil {
		t.Errorf("expected LinkUser match, got %q, %v", n, cl)
	}
}

func TestFindSuffixValueCompositeLit(t *testing.T) {
	if n, cl := FindSuffixValueCompositeLit(nil, "Cmd"); n != "" || cl != nil {
		t.Error("expected empty for nil decl")
	}
	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.TypeSpec{Name: dst.NewIdent("Ignored")},
			&dst.ValueSpec{},
			&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("Other")}, Values: []dst.Expr{&dst.CompositeLit{}}},
			&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("BadCmd")}, Values: []dst.Expr{dst.NewIdent("x")}},
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("GetUserCmd")},
				Values: []dst.Expr{&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{}}},
			},
		},
	}
	if n, cl := FindSuffixValueCompositeLit(decl, "Cmd"); n != "GetUserCmd" || cl == nil {
		t.Errorf("expected GetUserCmd, got %q, %v", n, cl)
	}
	if n, cl := FindSuffixValueCompositeLit(decl, "NonExistentSuffix"); n != "" || cl != nil {
		t.Errorf("expected empty for unmatched suffix, got %q, %v", n, cl)
	}
}

func TestUnwrapCompositeLit(t *testing.T) {
	if UnwrapCompositeLit(nil) != nil {
		t.Error("expected nil for nil expr")
	}
	if UnwrapCompositeLit(dst.NewIdent("x")) != nil {
		t.Error("expected nil for Ident")
	}
	if UnwrapCompositeLit(&dst.UnaryExpr{Op: token.AND, X: dst.NewIdent("x")}) != nil {
		t.Error("expected nil for &Ident")
	}
	if UnwrapCompositeLit(&dst.CompositeLit{}) == nil {
		t.Error("expected non-nil for CompositeLit")
	}
	if UnwrapCompositeLit(&dst.UnaryExpr{Op: token.AND, X: &dst.CompositeLit{}}) == nil {
		t.Error("expected non-nil for &CompositeLit")
	}
}

func TestStringLitAndKeyValues(t *testing.T) {
	if _, ok := StringLit(dst.NewIdent("x")); ok {
		t.Error("expected false for Ident")
	}
	if _, ok := StringLit(&dst.BasicLit{Kind: token.INT, Value: "123"}); ok {
		t.Error("expected false for INT BasicLit")
	}
	if val, ok := StringLit(&dst.BasicLit{Kind: token.STRING, Value: "unquoted_raw"}); !ok || val != "" {
		t.Errorf("expected empty string and true for malformed quote, got %q, %v", val, ok)
	}
	if val, ok := StringLit(&dst.BasicLit{Kind: token.STRING, Value: `"hello"`}); !ok || val != "hello" {
		t.Errorf("expected hello, got %q, %v", val, ok)
	}

	if _, _, ok := IdentKeyValue(dst.NewIdent("x")); ok {
		t.Error("expected false for non-KeyValueExpr")
	}
	if _, _, ok := IdentKeyValue(&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"k"`}, Value: dst.NewIdent("v")}); ok {
		t.Error("expected false for non-Ident key")
	}
	if k, v, ok := IdentKeyValue(&dst.KeyValueExpr{Key: dst.NewIdent("Title"), Value: dst.NewIdent("val")}); !ok || k != "Title" || v == nil {
		t.Errorf("unexpected IdentKeyValue result: %q, %v, %v", k, v, ok)
	}

	if _, _, ok := StringKeyValue(dst.NewIdent("x")); ok {
		t.Error("expected false for non-KeyValueExpr")
	}
	if _, _, ok := StringKeyValue(&dst.KeyValueExpr{Key: dst.NewIdent("k"), Value: dst.NewIdent("v")}); ok {
		t.Error("expected false for non-string key")
	}
	if k, v, ok := StringKeyValue(&dst.KeyValueExpr{Key: &dst.BasicLit{Kind: token.STRING, Value: `"http://cb"`}, Value: dst.NewIdent("v")}); !ok || k != "http://cb" || v == nil {
		t.Errorf("unexpected StringKeyValue result: %q, %v, %v", k, v, ok)
	}

	if LowerCamelTrimPrefix("Link", "Link") != "" {
		t.Error("expected empty string when only prefix present")
	}
	if LowerCamelTrimPrefix("LinkUserOrder", "Link") != "userOrder" {
		t.Errorf("expected userOrder, got %q", LowerCamelTrimPrefix("LinkUserOrder", "Link"))
	}
}
