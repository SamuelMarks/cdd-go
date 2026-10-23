// Package oauthflows provides parsing and emitting of OpenAPI OAuthFlows objects to and from Go AST composite literals.
package oauthflows

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI OAuthFlows object from a CompositeLit representation.
// Parameter cl is the AST composite literal, and it returns the parsed OAuthFlows or nil.
func Parse(cl *dst.CompositeLit) *openapi.OAuthFlows {
	if cl == nil {
		return nil
	}

	flows := &openapi.OAuthFlows{}
	hasContent := false

	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		valLit, ok := valExpr.(*dst.CompositeLit)
		if !ok {
			continue
		}
		flow := parseFlow(valLit)
		if flow == nil {
			continue
		}
		hasContent = true
		switch key {
		case "Implicit":
			flows.Implicit = flow
		case "Password":
			flows.Password = flow
		case "ClientCredentials":
			flows.ClientCredentials = flow
		case "AuthorizationCode":
			flows.AuthorizationCode = flow
		case "DeviceAuthorization":
			flows.DeviceAuthorization = flow
		}
	}

	if !hasContent {
		return nil
	}
	return flows
}

// parseFlow extracts a single OpenAPI OAuthFlow object from a dst.CompositeLit.
// Parameter cl is the flow AST composite literal, and it returns the parsed OAuthFlow or nil.
func parseFlow(cl *dst.CompositeLit) *openapi.OAuthFlow {
	flow := &openapi.OAuthFlow{
		Scopes: make(map[string]string),
	}
	hasContent := false

	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		if val, isStr := astutil.StringLit(valExpr); isStr {
			switch key {
			case "AuthorizationURL":
				flow.AuthorizationURL = val
				hasContent = true
			case "DeviceAuthorizationURL":
				flow.DeviceAuthorizationURL = val
				hasContent = true
			case "TokenURL":
				flow.TokenURL = val
				hasContent = true
			case "RefreshURL":
				flow.RefreshURL = val
				hasContent = true
			}
			continue
		}
		if mapLit, ok := valExpr.(*dst.CompositeLit); ok && key == "Scopes" {
			if parseScopesInto(mapLit, flow.Scopes) {
				hasContent = true
			}
		}
	}

	if !hasContent {
		return nil
	}
	return flow
}

// parseScopesInto populates scopes from a map CompositeLit and returns true if any valid entries were found.
// Parameter mapLit is the scopes map composite literal, and scopes is the destination map.
func parseScopesInto(mapLit *dst.CompositeLit, scopes map[string]string) bool {
	found := false
	for _, mapElt := range mapLit.Elts {
		k, vExpr, ok := astutil.StringKeyValue(mapElt)
		if !ok {
			continue
		}
		v, isStr := astutil.StringLit(vExpr)
		if !isStr {
			continue
		}
		scopes[k] = v
		found = true
	}
	return found
}
