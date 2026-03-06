package components

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/headers"
	"github.com/samuel/cdd-go/src/openapi"
	"github.com/samuel/cdd-go/src/parameters"
	"github.com/samuel/cdd-go/src/requestbodies"
	"github.com/samuel/cdd-go/src/responses"
	"github.com/samuel/cdd-go/src/schemas"
	"github.com/samuel/cdd-go/src/securityschemes"
)

// Emit formats an OpenAPI Components object into multiple struct/var representations.
func Emit(comp *openapi.Components) []dst.Decl {
	if comp == nil {
		return nil
	}

	var decls []dst.Decl

	// Security Schemes
	if comp.SecuritySchemes != nil {
		for name, scheme := range comp.SecuritySchemes {
			if decl := securityschemes.Emit(name, &scheme); decl != nil {
				decls = append(decls, decl)
			}
		}
	}

	// Reusable Parameters (emitted as variables)
	if comp.Parameters != nil {
		for name, param := range comp.Parameters {
			if pDecl := parameters.Emit(param); pDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.VAR,
					Specs: []dst.Spec{
						&dst.ValueSpec{
							Names: []*dst.Ident{dst.NewIdent("Param" + toPascalCase(name))},
							Type:  pDecl.Type,
						},
					},
				}
				if len(pDecl.Decs.Start) > 0 {
					decl.Decs.Start = pDecl.Decs.Start
				}
				decls = append(decls, decl)
			}
		}
	}

	// Reusable Headers
	if comp.Headers != nil {
		for name, header := range comp.Headers {
			if hDecl := headers.Emit(name, &header); hDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.VAR,
					Specs: []dst.Spec{
						&dst.ValueSpec{
							Names: []*dst.Ident{dst.NewIdent("Header" + toPascalCase(name))},
							Type:  hDecl.Type,
						},
					},
				}
				if len(hDecl.Decs.Start) > 0 {
					decl.Decs.Start = hDecl.Decs.Start
				}
				decls = append(decls, decl)
			}
		}
	}

	// Reusable Request Bodies
	if comp.RequestBodies != nil {
		for name, rb := range comp.RequestBodies {
			if rbDecl := requestbodies.Emit(&rb); rbDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.TYPE,
					Specs: []dst.Spec{
						&dst.TypeSpec{
							Name: dst.NewIdent("RequestBody" + toPascalCase(name)),
							Type: rbDecl.Type,
						},
					},
				}
				if len(rbDecl.Decs.Start) > 0 {
					decl.Decs.Start = rbDecl.Decs.Start
				}
				decls = append(decls, decl)
			}
		}
	}

	// Reusable Responses
	if comp.Responses != nil {
		for name, resp := range comp.Responses {
			resps := openapi.Responses{"200": resp}
			if respExprs := responses.Emit(resps); len(respExprs) > 0 {
				var rType dst.Expr = respExprs[0]
				if len(respExprs) > 2 {
					fields := []*dst.Field{}
					for i, expr := range respExprs {
						if ident, ok := expr.(*dst.Ident); ok && ident.Name == "error" {
							continue
						}
						fields = append(fields, &dst.Field{Type: expr, Names: []*dst.Ident{dst.NewIdent(fmt.Sprintf("F%d", i))}})
					}
					rType = &dst.StructType{Fields: &dst.FieldList{List: fields}}
				}

				decl := &dst.GenDecl{
					Tok: token.TYPE,
					Specs: []dst.Spec{
						&dst.TypeSpec{
							Name: dst.NewIdent("Response" + toPascalCase(name)),
							Type: rType,
						},
					},
				}
				decls = append(decls, decl)
			}
		}
	}

	// Schemas
	if comp.Schemas != nil {
		for name, schema := range comp.Schemas {
			if decl := schemas.Emit(name, &schema); decl != nil {
				decls = append(decls, decl)
			}
		}
	}

	return decls
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
