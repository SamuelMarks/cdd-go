package responses

import (
	"strings"

	"github.com/SamuelMarks/cdd-go/src/headers"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Responses object into a slice of Go return types.
func Emit(resps openapi.Responses) []dst.Expr {
	if resps == nil {
		return []dst.Expr{dst.NewIdent("error")}
	}

	var results []dst.Expr

	var successSchema *openapi.Schema
	var respHeaders map[string]openapi.Header
	for codeStr, resp := range resps {
		if strings.HasPrefix(codeStr, "2") {
			if resp.Content != nil {
				if mt, ok := resp.Content["application/json"]; ok && mt.Schema != nil {
					successSchema = mt.Schema
				}
			}
			if resp.Headers != nil && len(resp.Headers) > 0 {
				respHeaders = resp.Headers
			}
			break
		}
	}

	if successSchema != nil && successSchema.Ref != "" {
		parts := strings.Split(successSchema.Ref, "/")
		refName := parts[len(parts)-1]

		results = append(results, &dst.StarExpr{X: dst.NewIdent(refName)})
	}

	if respHeaders != nil {
		for hName, hDef := range respHeaders {
			hField := headers.Emit(hName, &hDef)
			if hField != nil {
				results = append(results, hField.Type)
			}
		}
	}

	results = append(results, dst.NewIdent("error"))
	return results
}
