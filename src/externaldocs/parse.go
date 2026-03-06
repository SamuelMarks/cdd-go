package externaldocs

import (
	"go/token"
	"strconv"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI ExternalDocs object from an AST representation.
func Parse(decl *dst.GenDecl) (*openapi.ExternalDocs, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && vs.Names[0].Name == "ExternalDocs" && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					docs := &openapi.ExternalDocs{}
					hasContent := false
					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyIdent, ok := kv.Key.(*dst.Ident); ok {
								if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
									val, _ := strconv.Unquote(valLit.Value)
									switch keyIdent.Name {
									case "Description":
										docs.Description = val
										hasContent = true
									case "URL":
										docs.URL = val
										hasContent = true
									}
								}
							}
						}
					}
					if hasContent {
						return docs, nil
					}
					return nil, nil
				}
			}
		}
	}

	return nil, nil
}
