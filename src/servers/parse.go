package servers

import (
	"go/token"
	"strconv"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Servers slice from an AST representation.
func Parse(decl *dst.GenDecl) ([]openapi.Server, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && vs.Names[0].Name == "Servers" && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					var servers []openapi.Server
					for _, elt := range cl.Elts {
						if srvLit, ok := elt.(*dst.CompositeLit); ok {
							srv := openapi.Server{}
							for _, srvElt := range srvLit.Elts {
								if kv, ok := srvElt.(*dst.KeyValueExpr); ok {
									if keyIdent, ok := kv.Key.(*dst.Ident); ok {
										if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
											val, _ := strconv.Unquote(valLit.Value)
											switch keyIdent.Name {
											case "URL":
												srv.URL = val
											case "Description":
												srv.Description = val
											}
										}
									}
								}
							}
							servers = append(servers, srv)
						}
					}
					return servers, nil
				}
			}
		}
	}

	return nil, nil
}
