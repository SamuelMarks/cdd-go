package links

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Link object from a struct representation.
func Parse(decl *dst.GenDecl) (string, *openapi.Link) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && strings.HasPrefix(vs.Names[0].Name, "Link") && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					name := strings.TrimPrefix(vs.Names[0].Name, "Link")
					if name != "" {
						name = strings.ToLower(name[:1]) + name[1:]
					}

					link := &openapi.Link{}
					hasContent := false

					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyIdent, ok := kv.Key.(*dst.Ident); ok {
								if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
									val, _ := strconv.Unquote(valLit.Value)
									switch keyIdent.Name {
									case "OperationRef":
										link.OperationRef = val
										hasContent = true
									case "OperationID":
										link.OperationID = val
										hasContent = true
									case "Description":
										link.Description = val
										hasContent = true
									}
								}
							}
						}
					}

					if hasContent {
						return name, link
					}
				}
			}
		}
	}

	return "", nil
}
