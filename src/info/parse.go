package info

import (
	"go/token"
	"strconv"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// Parse extracts an OpenAPI Info object from an AST representation.
func Parse(decl *dst.GenDecl) (*openapi.Info, error) {
	if decl == nil || decl.Tok != token.CONST {
		return nil, nil
	}

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && vs.Names[0].Name == "Info" && len(vs.Values) > 0 {
				if cl, ok := vs.Values[0].(*dst.CompositeLit); ok {
					info := &openapi.Info{}
					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyIdent, ok := kv.Key.(*dst.Ident); ok {
								if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
									val, _ := strconv.Unquote(valLit.Value)
									switch keyIdent.Name {
									case "Title":
										info.Title = val
									case "Version":
										info.Version = val
									case "Description":
										info.Description = val
									case "Summary":
										info.Summary = val
									case "TermsOfService":
										info.TermsOfService = val
									}
								} else if childCl, ok := kv.Value.(*dst.CompositeLit); ok {
									if keyIdent.Name == "Contact" {
										info.Contact = &openapi.Contact{}
										for _, childElt := range childCl.Elts {
											if childKv, ok := childElt.(*dst.KeyValueExpr); ok {
												if childKeyIdent, ok := childKv.Key.(*dst.Ident); ok {
													if childValLit, ok := childKv.Value.(*dst.BasicLit); ok && childValLit.Kind == token.STRING {
														val, _ := strconv.Unquote(childValLit.Value)
														switch childKeyIdent.Name {
														case "Name":
															info.Contact.Name = val
														case "URL":
															info.Contact.URL = val
														case "Email":
															info.Contact.Email = val
														}
													}
												}
											}
										}
									} else if keyIdent.Name == "License" {
										info.License = &openapi.License{}
										for _, childElt := range childCl.Elts {
											if childKv, ok := childElt.(*dst.KeyValueExpr); ok {
												if childKeyIdent, ok := childKv.Key.(*dst.Ident); ok {
													if childValLit, ok := childKv.Value.(*dst.BasicLit); ok && childValLit.Kind == token.STRING {
														val, _ := strconv.Unquote(childValLit.Value)
														switch childKeyIdent.Name {
														case "Name":
															info.License.Name = val
														case "URL":
															info.License.URL = val
														case "Identifier":
															info.License.Identifier = val
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
					return info, nil
				}
			}
		}
	}

	return nil, nil
}
