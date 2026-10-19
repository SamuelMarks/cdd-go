// Package webhooks provides parsing and emitting of OpenAPI Webhooks maps to and from Go AST declarations.
package webhooks

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Webhooks object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the webhook name and parsed PathItem map.
func Parse(decl *dst.GenDecl) (string, map[string]openapi.PathItem) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	rawName, cl := astutil.FindValueCompositeLit(decl, "", "Webhook")
	if cl == nil {
		return "", nil
	}

	name := astutil.LowerCamelTrimPrefix(rawName, "Webhook")
	webhooks := make(map[string]openapi.PathItem)
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
		pi, ok := parseWebhookPathItem(valCl)
		if ok {
			hasContent = true
		}
		webhooks[keyVal] = pi
	}

	if !hasContent {
		return "", nil
	}
	return name, webhooks
}

// parseWebhookPathItem extracts a PathItem from a CompositeLit inside a Webhooks map.
// Parameter valCl is the PathItem AST composite literal, and it returns the PathItem and whether content was found.
func parseWebhookPathItem(valCl *dst.CompositeLit) (openapi.PathItem, bool) {
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
