package webhooks

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Webhooks object from a struct representation.
func Parse(decl *dst.GenDecl) (string, map[string]openapi.PathItem) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && strings.HasPrefix(vs.Names[0].Name, "Webhook") && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					name := strings.TrimPrefix(vs.Names[0].Name, "Webhook")
					if name != "" {
						name = strings.ToLower(name[:1]) + name[1:]
					}

					webhooks := make(map[string]openapi.PathItem)
					hasContent := false

					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyLit, ok := kv.Key.(*dst.BasicLit); ok && keyLit.Kind == token.STRING {
								keyVal, _ := strconv.Unquote(keyLit.Value)
								if valCl, ok := kv.Value.(*dst.CompositeLit); ok {
									pi := openapi.PathItem{}
									for _, piElt := range valCl.Elts {
										if piKv, ok := piElt.(*dst.KeyValueExpr); ok {
											if piKey, ok := piKv.Key.(*dst.Ident); ok && piKey.Name == "Summary" {
												if piVal, ok := piKv.Value.(*dst.BasicLit); ok && piVal.Kind == token.STRING {
													pi.Summary, _ = strconv.Unquote(piVal.Value)
													hasContent = true
												}
											}
										}
									}
									webhooks[keyVal] = pi
								}
							}
						}
					}

					if hasContent {
						return name, webhooks
					}
				}
			}
		}
	}

	return "", nil
}
