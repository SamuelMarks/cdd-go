package components

import (
	"go/token"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/headers"
	"github.com/samuel/cdd-go/src/openapi"
	"github.com/samuel/cdd-go/src/parameters"
	"github.com/samuel/cdd-go/src/requestbodies"
	"github.com/samuel/cdd-go/src/responses"
	"github.com/samuel/cdd-go/src/schemas"
	"github.com/samuel/cdd-go/src/securityschemes"
)

// Parse extracts OpenAPI components from a file
func Parse(file *dst.File) *openapi.Components {
	if file == nil {
		return nil
	}

	comp := &openapi.Components{
		Schemas:         make(map[string]openapi.Schema),
		SecuritySchemes: make(map[string]openapi.SecurityScheme),
		Parameters:      make(map[string]openapi.Parameter),
		Headers:         make(map[string]openapi.Header),
		RequestBodies:   make(map[string]openapi.RequestBody),
		Responses:       make(map[string]openapi.Response),
	}

	hasContent := false

	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok {
			if genDecl.Tok == token.VAR {
				for _, spec := range genDecl.Specs {
					if vs, ok := spec.(*dst.ValueSpec); ok && len(vs.Names) > 0 {
						name := vs.Names[0].Name

						if strings.HasPrefix(name, "SecurityScheme") {
							sn, scheme := securityschemes.Parse(genDecl)
							if scheme != nil {
								comp.SecuritySchemes[sn] = *scheme
								hasContent = true
							}
						} else if strings.HasPrefix(name, "Param") {
							f := &dst.Field{
								Names: []*dst.Ident{dst.NewIdent(strings.TrimPrefix(name, "Param"))},
								Type:  vs.Type,
							}
							f.Decs.Start = genDecl.Decs.Start
							if p := parameters.Parse(f); p != nil {
								pName := strings.TrimPrefix(name, "Param")
								pName = strings.ToLower(pName[:1]) + pName[1:]
								p.Name = pName
								comp.Parameters[pName] = *p
								hasContent = true
							}
						} else if strings.HasPrefix(name, "Header") {
							f := &dst.Field{
								Names: []*dst.Ident{dst.NewIdent(strings.TrimPrefix(name, "Header"))},
								Type:  vs.Type,
							}
							f.Decs.Start = genDecl.Decs.Start
							if h := headers.Parse(f); h != nil {
								hName := strings.TrimPrefix(name, "Header")
								hName = strings.ToLower(hName[:1]) + hName[1:]
								comp.Headers[hName] = *h
								hasContent = true
							}
						}
					}
				}
			} else if genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if ts, ok := spec.(*dst.TypeSpec); ok {
						name := ts.Name.Name
						if strings.HasPrefix(name, "RequestBody") {
							f := &dst.Field{Type: ts.Type}
							f.Decs.Start = genDecl.Decs.Start
							if rb := requestbodies.Parse(f); rb != nil {
								rbName := strings.TrimPrefix(name, "RequestBody")
								rbName = strings.ToLower(rbName[:1]) + rbName[1:]
								comp.RequestBodies[rbName] = *rb
								hasContent = true
							}
						} else if strings.HasPrefix(name, "Response") {
							fl := &dst.FieldList{List: []*dst.Field{{Type: ts.Type}}}
							if st, ok := ts.Type.(*dst.StructType); ok {
								fl = st.Fields
							}
							if resps := responses.Parse(fl); resps != nil {
								if r, ok := resps["200"]; ok {
									rName := strings.TrimPrefix(name, "Response")
									rName = strings.ToLower(rName[:1]) + rName[1:]
									comp.Responses[rName] = r
									hasContent = true
								}
							}
						} else {
							sn, s := schemas.Parse(genDecl)
							if s != nil && sn != "" {
								comp.Schemas[sn] = *s
								hasContent = true
							}
						}
					}
				}
			}
		}
	}

	if !hasContent {
		return nil
	}

	return comp
}
