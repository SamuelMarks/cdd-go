package oauthflows

import (
	"go/token"
	"strconv"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI OAuthFlows object from a CompositeLit representation.
func Parse(cl *dst.CompositeLit) *openapi.OAuthFlows {
	if cl == nil {
		return nil
	}

	flows := &openapi.OAuthFlows{}
	hasContent := false

	for _, elt := range cl.Elts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if keyIdent, ok := kv.Key.(*dst.Ident); ok {
				if valLit, ok := kv.Value.(*dst.CompositeLit); ok {
					flow := parseFlow(valLit)
					if flow != nil {
						hasContent = true
						switch keyIdent.Name {
						case "Implicit":
							flows.Implicit = flow
						case "Password":
							flows.Password = flow
						case "ClientCredentials":
							flows.ClientCredentials = flow
						case "AuthorizationCode":
							flows.AuthorizationCode = flow
						}
					}
				}
			}
		}
	}

	if !hasContent {
		return nil
	}

	return flows
}

func parseFlow(cl *dst.CompositeLit) *openapi.OAuthFlow {
	flow := &openapi.OAuthFlow{
		Scopes: make(map[string]string),
	}
	hasContent := false

	for _, elt := range cl.Elts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if keyIdent, ok := kv.Key.(*dst.Ident); ok {
				if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
					val, _ := strconv.Unquote(valLit.Value)
					switch keyIdent.Name {
					case "AuthorizationURL":
						flow.AuthorizationURL = val
						hasContent = true
					case "TokenURL":
						flow.TokenURL = val
						hasContent = true
					case "RefreshURL":
						flow.RefreshURL = val
						hasContent = true
					}
				} else if mapLit, ok := kv.Value.(*dst.CompositeLit); ok && keyIdent.Name == "Scopes" {
					for _, mapElt := range mapLit.Elts {
						if mapKV, ok := mapElt.(*dst.KeyValueExpr); ok {
							kStr, kOk := mapKV.Key.(*dst.BasicLit)
							vStr, vOk := mapKV.Value.(*dst.BasicLit)
							if kOk && vOk && kStr.Kind == token.STRING && vStr.Kind == token.STRING {
								k, _ := strconv.Unquote(kStr.Value)
								v, _ := strconv.Unquote(vStr.Value)
								flow.Scopes[k] = v
								hasContent = true
							}
						}
					}
				}
			}
		}
	}

	if !hasContent {
		return nil
	}

	return flow
}
