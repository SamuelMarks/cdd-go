package securityschemes

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/oauthflows"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI SecurityScheme object from a struct representation.
func Parse(decl *dst.GenDecl) (string, *openapi.SecurityScheme) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && strings.HasPrefix(vs.Names[0].Name, "SecurityScheme") && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					name := strings.TrimPrefix(vs.Names[0].Name, "SecurityScheme")
					if name != "" {
						name = strings.ToLower(name[:1]) + name[1:]
					}

					scheme := &openapi.SecurityScheme{}
					hasContent := false

					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyIdent, ok := kv.Key.(*dst.Ident); ok {
								if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
									val, _ := strconv.Unquote(valLit.Value)
									switch keyIdent.Name {
									case "Type":
										scheme.Type = val
										hasContent = true
									case "Description":
										scheme.Description = val
										hasContent = true
									case "Name":
										scheme.Name = val
										hasContent = true
									case "In":
										scheme.In = val
										hasContent = true
									case "Scheme":
										scheme.Scheme = val
										hasContent = true
									case "BearerFormat":
										scheme.BearerFormat = val
										hasContent = true
									case "OpenIDConnectURL":
										scheme.OpenIDConnectURL = val
										hasContent = true
									}
								} else if clLit, ok := kv.Value.(*dst.CompositeLit); ok && keyIdent.Name == "Flows" {
									flows := oauthflows.Parse(clLit)
									if flows != nil {
										scheme.Flows = flows
										hasContent = true
									}
								}
							}
						}
					}

					if hasContent {
						return name, scheme
					}
				}
			}
		}
	}

	return "", nil
}
