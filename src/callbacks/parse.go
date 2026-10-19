// Package callbacks provides parsing and emitting of OpenAPI Callback objects to and from Go AST declarations.
package callbacks

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Callback object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the callback name and parsed Callback object.
func Parse(decl *dst.GenDecl) (string, *openapi.Callback) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	rawName, cl := astutil.FindValueCompositeLit(decl, "", "Callback")
	if cl == nil {
		return "", nil
	}

	name := astutil.LowerCamelTrimPrefix(rawName, "Callback")
	callback := make(openapi.Callback)
	hasContent := false

	for _, elt := range cl.Elts {
		keyVal, valExpr, ok := astutil.StringKeyValue(elt)
		if !ok {
			continue
		}
		valCl, ok := valExpr.(*dst.CompositeLit)
		if !ok {
			continue
		}
		pi, ok := parseCallbackPathItem(valCl)
		if ok {
			hasContent = true
		}
		callback[keyVal] = pi
	}

	if !hasContent {
		return "", nil
	}
	return name, &callback
}

// parseCallbackPathItem extracts a PathItem from a CompositeLit inside a Callback map.
// Parameter valCl is the PathItem AST composite literal, and it returns the PathItem and whether content was found.
func parseCallbackPathItem(valCl *dst.CompositeLit) (openapi.PathItem, bool) {
	pi := openapi.PathItem{}
	found := false
	for _, piElt := range valCl.Elts {
		k, vExpr, ok := astutil.IdentKeyValue(piElt)
		if !ok || k != "Summary" {
			continue
		}
		if strVal, isStr := astutil.StringLit(vExpr); isStr {
			pi.Summary = strVal
			found = true
		}
	}
	return pi, found
}
