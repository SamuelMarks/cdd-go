// Package astutil provides shared, strongly-typed helpers for inspecting and unwrapping Go DST AST nodes without deep nesting.
package astutil

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/dave/dst"
)

// FindValueCompositeLit locates the first ValueSpec in decl whose identifier matches exactName (or starts with prefix when exactName is empty) and returns its identifier name and CompositeLit value.
// Parameter decl is the generic AST declaration, exactName is an optional exact variable name match, and prefix is an optional variable name prefix.
func FindValueCompositeLit(decl *dst.GenDecl, exactName string, prefix string) (string, *dst.CompositeLit) {
	if decl == nil || (decl.Tok != token.VAR && decl.Tok != token.CONST) {
		return "", nil
	}
	for _, spec := range decl.Specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
			continue
		}
		name := vs.Names[0].Name
		if exactName != "" && name != exactName {
			continue
		}
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		if cl := UnwrapCompositeLit(vs.Values[0]); cl != nil {
			return name, cl
		}
	}
	return "", nil
}

// FindSuffixValueCompositeLit locates the first ValueSpec in a VAR declaration whose identifier ends with suffix and returns its identifier name and CompositeLit value.
// Parameter decl is the generic AST declaration, and suffix is the required identifier suffix.
func FindSuffixValueCompositeLit(decl *dst.GenDecl, suffix string) (string, *dst.CompositeLit) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	for _, spec := range decl.Specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
			continue
		}
		name := vs.Names[0].Name
		if !strings.HasSuffix(name, suffix) {
			continue
		}
		if cl := UnwrapCompositeLit(vs.Values[0]); cl != nil {
			return name, cl
		}
	}
	return "", nil
}

// UnwrapCompositeLit extracts a *dst.CompositeLit from either a direct CompositeLit or an address-of (&CompositeLit) UnaryExpr.
// Parameter expr is the AST expression to unwrap, and it returns the underlying CompositeLit or nil.
func UnwrapCompositeLit(expr dst.Expr) *dst.CompositeLit {
	if expr == nil {
		return nil
	}
	if cl, ok := expr.(*dst.CompositeLit); ok {
		return cl
	}
	if unary, ok := expr.(*dst.UnaryExpr); ok && unary.Op == token.AND {
		if cl, ok := unary.X.(*dst.CompositeLit); ok {
			return cl
		}
	}
	return nil
}

// StringLit extracts and unquotes a string literal from expr if it is a token.STRING BasicLit.
// Parameter expr is the AST expression, and it returns the unquoted string and true if expr is a string literal.
func StringLit(expr dst.Expr) (string, bool) {
	lit, ok := expr.(*dst.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", true
	}
	return val, true
}

// IdentKeyValue extracts the identifier key name and value expression from elt if it is a KeyValueExpr with an Ident key.
// Parameter elt is the composite literal element expression, and it returns the key name, value expression, and true on match.
func IdentKeyValue(elt dst.Expr) (string, dst.Expr, bool) {
	kv, ok := elt.(*dst.KeyValueExpr)
	if !ok {
		return "", nil, false
	}
	ident, ok := kv.Key.(*dst.Ident)
	if !ok {
		return "", nil, false
	}
	return ident.Name, kv.Value, true
}

// StringKeyValue extracts an unquoted string key and value expression from elt if it is a KeyValueExpr with a string literal key.
// Parameter elt is the composite literal element expression, and it returns the unquoted key string, value expression, and true on match.
func StringKeyValue(elt dst.Expr) (string, dst.Expr, bool) {
	kv, ok := elt.(*dst.KeyValueExpr)
	if !ok {
		return "", nil, false
	}
	keyStr, ok := StringLit(kv.Key)
	if !ok {
		return "", nil, false
	}
	return keyStr, kv.Value, true
}

// LowerCamelTrimPrefix removes prefix from name and lowercases the first character of the remaining identifier.
// Parameter name is the full identifier, and prefix is the prefix to strip.
func LowerCamelTrimPrefix(name, prefix string) string {
	trimmed := strings.TrimPrefix(name, prefix)
	if trimmed == "" {
		return ""
	}
	return strings.ToLower(trimmed[:1]) + trimmed[1:]
}
