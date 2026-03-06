package tags

import (
	"go/token"
	"strconv"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI Tags slice from an AST representation.
func Parse(decl *dst.GenDecl) ([]openapi.Tag, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && vs.Names[0].Name == "Tags" && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					var tags []openapi.Tag
					for _, elt := range cl.Elts {
						if tagLit, ok := elt.(*dst.CompositeLit); ok {
							tag := openapi.Tag{}
							for _, tagElt := range tagLit.Elts {
								if kv, ok := tagElt.(*dst.KeyValueExpr); ok {
									if keyIdent, ok := kv.Key.(*dst.Ident); ok {
										if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
											val, _ := strconv.Unquote(valLit.Value)
											switch keyIdent.Name {
											case "Name":
												tag.Name = val
											case "Description":
												tag.Description = val
											}
										}
									}
								}
							}
							tags = append(tags, tag)
						}
					}
					return tags, nil
				}
			}
		}
	}

	return nil, nil
}
