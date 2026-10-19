// Package info provides parsing and emitting of OpenAPI Info metadata to and from Go AST declarations.
package info

import (
	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Info object from an AST representation.
// Parameter decl is the AST generic declaration, and it returns the parsed OpenAPI Info and error.
func Parse(decl *dst.GenDecl) (*openapi.Info, error) {
	_, cl := astutil.FindValueCompositeLit(decl, "Info", "")
	if cl == nil {
		return nil, nil
	}

	info := &openapi.Info{}
	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		if val, isStr := astutil.StringLit(valExpr); isStr {
			switch key {
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
			continue
		}
		childCl, ok := valExpr.(*dst.CompositeLit)
		if !ok {
			continue
		}
		switch key {
		case "Contact":
			info.Contact = parseContact(childCl)
		case "License":
			info.License = parseLicense(childCl)
		}
	}
	return info, nil
}

// parseContact extracts an OpenAPI Contact object from a CompositeLit.
// Parameter cl is the contact AST composite literal, and it returns the parsed Contact.
func parseContact(cl *dst.CompositeLit) *openapi.Contact {
	c := &openapi.Contact{}
	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		val, isStr := astutil.StringLit(valExpr)
		if !isStr {
			continue
		}
		switch key {
		case "Name":
			c.Name = val
		case "URL":
			c.URL = val
		case "Email":
			c.Email = val
		}
	}
	return c
}

// parseLicense extracts an OpenAPI License object from a CompositeLit.
// Parameter cl is the license AST composite literal, and it returns the parsed License.
func parseLicense(cl *dst.CompositeLit) *openapi.License {
	l := &openapi.License{}
	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		val, isStr := astutil.StringLit(valExpr)
		if !isStr {
			continue
		}
		switch key {
		case "Name":
			l.Name = val
		case "URL":
			l.URL = val
		case "Identifier":
			l.Identifier = val
		}
	}
	return l
}
