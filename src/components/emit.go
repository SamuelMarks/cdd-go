// Package components provides parsing and emitting of OpenAPI Components to and from Go AST declarations.
package components

import (
	"fmt"
	"go/token"
	"sort"

	"github.com/SamuelMarks/cdd-go/src/headers"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/oauthflows"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/parameters"
	"github.com/SamuelMarks/cdd-go/src/requestbodies"
	"github.com/SamuelMarks/cdd-go/src/responses"
	"github.com/SamuelMarks/cdd-go/src/schemas"
	"github.com/SamuelMarks/cdd-go/src/securityschemes"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Components object into multiple struct/var representations.
func Emit(comp *openapi.Components) []dst.Decl {
	if comp == nil {
		return nil
	}

	var decls []dst.Decl

	// Security Schemes
	if comp.SecuritySchemes != nil {
		keys := make([]string, 0, len(comp.SecuritySchemes))
		for k := range comp.SecuritySchemes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		needsOAuthTypes := false
		for _, name := range keys {
			scheme := comp.SecuritySchemes[name]
			if scheme.Type == "oauth2" || scheme.Flows != nil {
				needsOAuthTypes = true
			}
		}
		if needsOAuthTypes {
			decls = append(decls, oauthflows.EmitTypes()...)
		}
		for _, name := range keys {
			scheme := comp.SecuritySchemes[name]
			if decl := securityschemes.Emit(name, &scheme); decl != nil {
				decls = append(decls, decl)
			}
		}
	}

	// Reusable Parameters (emitted as variables)
	if comp.Parameters != nil {
		keys := make([]string, 0, len(comp.Parameters))
		for k := range comp.Parameters {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			param := comp.Parameters[name]
			if pDecl := parameters.Emit(param); pDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.VAR,
					Specs: []dst.Spec{
						&dst.ValueSpec{
							Names: []*dst.Ident{dst.NewIdent("Param" + naming.ToPascalCase(name))},
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
		keys := make([]string, 0, len(comp.Headers))
		for k := range comp.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			header := comp.Headers[name]
			if hDecl := headers.Emit(name, &header); hDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.VAR,
					Specs: []dst.Spec{
						&dst.ValueSpec{
							Names: []*dst.Ident{dst.NewIdent("Header" + naming.ToPascalCase(name))},
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
		keys := make([]string, 0, len(comp.RequestBodies))
		for k := range comp.RequestBodies {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			rb := comp.RequestBodies[name]
			if rbDecl := requestbodies.Emit(&rb); rbDecl != nil {
				decl := &dst.GenDecl{
					Tok: token.TYPE,
					Specs: []dst.Spec{
						&dst.TypeSpec{
							Name: dst.NewIdent("RequestBody" + naming.ToPascalCase(name)),
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
		keys := make([]string, 0, len(comp.Responses))
		for k := range comp.Responses {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			resp := comp.Responses[name]
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
							Name: dst.NewIdent("Response" + naming.ToPascalCase(name)),
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
		keys := make([]string, 0, len(comp.Schemas))
		for k := range comp.Schemas {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			schema := comp.Schemas[name]
			if decl, err := schemas.Emit(name, &schema); err == nil && decl != nil {
				decls = append(decls, decl)
			}
		}
	}

	return decls
}
