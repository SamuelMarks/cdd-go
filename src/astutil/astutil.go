// Package astutil provides shared, strongly-typed helpers for inspecting and unwrapping Go DST AST nodes without deep nesting.
package astutil

import (
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
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

// EmitPathItemMapDecl emits a Go AST VAR declaration representing a named map of OpenAPI PathItems (used by callbacks and webhooks).
// Parameter typeName is the composite literal type name, varPrefix is the variable name prefix, name is the item name, and items is the PathItem map.
func EmitPathItemMapDecl(typeName, varPrefix, name string, items map[string]openapi.PathItem) *dst.GenDecl {
	if items == nil {
		return nil
	}

	cl := &dst.CompositeLit{
		Type: dst.NewIdent(typeName),
		Elts: []dst.Expr{},
	}

	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := items[k]
		piElts := []dst.Expr{
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Summary"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", v.Summary)},
			},
		}
		if v.Description != "" {
			piElts = append(piElts, &dst.KeyValueExpr{
				Key:   dst.NewIdent("Description"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", v.Description)},
			})
		}
		for _, mo := range openapi.PathItemOperations(&v) {
			fieldKey := strings.ToUpper(mo.Method[:1]) + mo.Method[1:]
			piElts = append(piElts, &dst.KeyValueExpr{
				Key:   dst.NewIdent(fieldKey),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", mo.Operation.OperationID)},
			})
		}
		valLit := &dst.CompositeLit{
			Type: dst.NewIdent("PathItem"),
			Elts: piElts,
		}
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", k)},
			Value: valLit,
		})
	}

	return &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent(varPrefix + naming.ToPascalCase(name))},
				Values: []dst.Expr{cl},
			},
		},
	}
}

// ParsePathItemMapDecl extracts a named map of OpenAPI PathItems from a VAR GenDecl (used by callbacks and webhooks).
// Parameter decl is the generic AST declaration, and varPrefix is the variable name prefix ("Callback" or "Webhook").
func ParsePathItemMapDecl(decl *dst.GenDecl, varPrefix string) (string, map[string]openapi.PathItem) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	rawName, cl := FindValueCompositeLit(decl, "", varPrefix)
	if cl == nil {
		return "", nil
	}

	name := LowerCamelTrimPrefix(rawName, varPrefix)
	result := make(map[string]openapi.PathItem)
	hasContent := false

	for _, elt := range cl.Elts {
		keyVal, valExpr, ok := StringKeyValue(elt)
		if !ok {
			continue
		}
		valCl, ok := valExpr.(*dst.CompositeLit)
		if !ok {
			continue
		}
		pi, found := parsePathItemCompositeLit(valCl)
		if found {
			hasContent = true
		}
		result[keyVal] = pi
	}

	if !hasContent {
		return "", nil
	}
	return name, result
}

// parsePathItemCompositeLit reconstructs an openapi.PathItem from its AST CompositeLit representation.
// Parameter valCl is the PathItem composite literal, and it returns the parsed PathItem and true if any field was populated.
func parsePathItemCompositeLit(valCl *dst.CompositeLit) (openapi.PathItem, bool) {
	pi := openapi.PathItem{}
	found := false
	for _, piElt := range valCl.Elts {
		k, vExpr, ok := IdentKeyValue(piElt)
		if !ok {
			continue
		}
		strVal, isStr := StringLit(vExpr)
		if !isStr {
			continue
		}
		switch k {
		case "Summary":
			pi.Summary = strVal
			found = true
		case "Description":
			pi.Description = strVal
			found = true
		case "Get", "Post", "Put", "Delete", "Patch", "Options", "Head", "Trace":
			pi.SetOperation(k, &openapi.Operation{OperationID: strVal})
			found = true
		}
	}
	return pi, found
}
