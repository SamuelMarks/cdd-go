// Package components provides parsing and emitting of OpenAPI Components to and from Go AST declarations.
package components

import (
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/headers"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/parameters"
	"github.com/SamuelMarks/cdd-go/src/requestbodies"
	"github.com/SamuelMarks/cdd-go/src/responses"
	"github.com/SamuelMarks/cdd-go/src/schemas"
	"github.com/SamuelMarks/cdd-go/src/securityschemes"
	"github.com/dave/dst"
)

// Parse extracts OpenAPI components from a file.
// Parameter file is the Go AST file, and it returns the extracted Components or nil if empty.
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
		genDecl, ok := decl.(*dst.GenDecl)
		if !ok {
			continue
		}
		switch genDecl.Tok {
		case token.VAR:
			if parseVarComponents(genDecl, comp) {
				hasContent = true
			}
		case token.TYPE:
			if parseTypeComponents(genDecl, comp) {
				hasContent = true
			}
		}
	}

	if !hasContent {
		return nil
	}
	return comp
}

// filterComponentDecs removes auto-generated fallback GoDoc lines from decs.
// Parameter decs is the input AST decorations, name is the symbol identifier, and it returns the filtered decorations.
func filterComponentDecs(decs dst.Decorations, name string) dst.Decorations {
	var out dst.Decorations
	autoType := "// " + name + " represents the " + name + " type."
	autoVal := "// " + name + " holds the " + name + " value."
	for _, d := range decs {
		trimmed := strings.TrimSpace(d)
		if trimmed == autoType || trimmed == autoVal {
			continue
		}
		out = append(out, d)
	}
	return out
}

// parseVarComponents parses variable declarations (SecuritySchemes, Parameters, Headers) into comp.
// Parameter genDecl is the VAR declaration, and comp is the target Components struct.
func parseVarComponents(genDecl *dst.GenDecl, comp *openapi.Components) bool {
	found := false
	for _, spec := range genDecl.Specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok || len(vs.Names) == 0 {
			continue
		}
		name := vs.Names[0].Name
		cleanDecs := filterComponentDecs(genDecl.Decs.Start, name)
		switch {
		case strings.HasPrefix(name, "SecurityScheme"):
			singleDecl := &dst.GenDecl{Tok: token.VAR, Specs: []dst.Spec{vs}, Decs: genDecl.Decs}
			if sn, scheme := securityschemes.Parse(singleDecl); scheme != nil {
				comp.SecuritySchemes[sn] = *scheme
				found = true
			}
		case strings.HasPrefix(name, "Param"):
			f := &dst.Field{
				Names: []*dst.Ident{dst.NewIdent(strings.TrimPrefix(name, "Param"))},
				Type:  vs.Type,
			}
			f.Decs.Start = cleanDecs
			if p := parameters.Parse(f); p != nil {
				pName := astutil.LowerCamelTrimPrefix(name, "Param")
				p.Name = pName
				comp.Parameters[pName] = *p
				found = true
			}
		case strings.HasPrefix(name, "Header"):
			f := &dst.Field{
				Names: []*dst.Ident{dst.NewIdent(strings.TrimPrefix(name, "Header"))},
				Type:  vs.Type,
			}
			f.Decs.Start = cleanDecs
			if h := headers.Parse(f); h != nil {
				hName := astutil.LowerCamelTrimPrefix(name, "Header")
				comp.Headers[hName] = *h
				found = true
			}
		}
	}
	return found
}

// parseTypeComponents parses type declarations (RequestBodies, Responses, Schemas) into comp.
// Parameter genDecl is the TYPE declaration, and comp is the target Components struct.
func parseTypeComponents(genDecl *dst.GenDecl, comp *openapi.Components) bool {
	found := false
	for _, spec := range genDecl.Specs {
		ts, ok := spec.(*dst.TypeSpec)
		if !ok {
			continue
		}
		name := ts.Name.Name
		cleanDecs := filterComponentDecs(genDecl.Decs.Start, name)
		switch {
		case name == "OAuthFlow" || name == "OAuthFlows":
			continue
		case strings.HasPrefix(name, "RequestBody"):
			f := &dst.Field{Type: ts.Type}
			f.Decs.Start = cleanDecs
			if rb := requestbodies.Parse(f); rb != nil {
				rbName := astutil.LowerCamelTrimPrefix(name, "RequestBody")
				comp.RequestBodies[rbName] = *rb
				found = true
			}
		case strings.HasPrefix(name, "Response"):
			fl := &dst.FieldList{List: []*dst.Field{{Type: ts.Type}}}
			if st, ok := ts.Type.(*dst.StructType); ok {
				fl = st.Fields
			}
			if resps := responses.Parse(fl); resps != nil {
				if r, ok := resps["200"]; ok {
					rName := astutil.LowerCamelTrimPrefix(name, "Response")
					comp.Responses[rName] = r
					found = true
				}
			}
		default:
			singleDecs := genDecl.Decs
			singleDecs.Start = cleanDecs
			singleDecl := &dst.GenDecl{Tok: token.TYPE, Specs: []dst.Spec{ts}, Decs: singleDecs}
			if sn, s, err := schemas.Parse(singleDecl); err == nil && s != nil && sn != "" {
				comp.Schemas[sn] = *s
				found = true
			}
		}
	}
	return found
}
