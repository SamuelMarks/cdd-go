// Package links provides parsing and emitting of OpenAPI Link objects to and from Go AST declarations.
package links

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Link object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the link name and parsed Link object.
func Parse(decl *dst.GenDecl) (string, *openapi.Link) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	rawName, cl := astutil.FindValueCompositeLit(decl, "", "Link")
	if cl == nil {
		return "", nil
	}

	name := astutil.LowerCamelTrimPrefix(rawName, "Link")
	link := &openapi.Link{}
	hasContent := false

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

	if !hasContent {
		return "", nil
	}
	return name, link
}
