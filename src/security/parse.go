package security

import (
	"go/token"
	"strconv"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI SecurityRequirement slice from an AST representation.
func Parse(expr dst.Expr) []openapi.SecurityRequirement {
	if expr == nil {
		return nil
	}

	cl, ok := expr.(*dst.CompositeLit)
	if !ok {
		return nil
	}

	var security []openapi.SecurityRequirement

	for _, elt := range cl.Elts {
		mapLit, ok := elt.(*dst.CompositeLit)
		if !ok {
			continue
		}

		secReq := make(openapi.SecurityRequirement)
		for _, mapElt := range mapLit.Elts {
			kv, ok := mapElt.(*dst.KeyValueExpr)
			if !ok {
				continue
			}

			keyLit, ok := kv.Key.(*dst.BasicLit)
			if !ok || keyLit.Kind != token.STRING {
				continue
			}

			key, err := strconv.Unquote(keyLit.Value)
			if err != nil {
				key = keyLit.Value
			}

			valLit, ok := kv.Value.(*dst.CompositeLit)
			if !ok {
				continue
			}

			var scopes []string
			for _, scopeElt := range valLit.Elts {
				if scopeStr, ok := scopeElt.(*dst.BasicLit); ok && scopeStr.Kind == token.STRING {
					s, err := strconv.Unquote(scopeStr.Value)
					if err != nil {
						s = scopeStr.Value
					}
					scopes = append(scopes, s)
				}
			}

			secReq[key] = scopes
		}

		if len(secReq) > 0 {
			security = append(security, secReq)
		}
	}

	if len(security) == 0 {
		return nil
	}

	return security
}
