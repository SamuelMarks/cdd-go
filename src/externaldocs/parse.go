// Package externaldocs provides parsing and emitting of OpenAPI ExternalDocs objects to and from Go AST declarations.
package externaldocs

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI ExternalDocs object from an AST representation.
// Parameter decl is the generic AST declaration, and it returns the parsed ExternalDocs and error.
func Parse(decl *dst.GenDecl) (*openapi.ExternalDocs, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}
	_, cl := astutil.FindValueCompositeLit(decl, "ExternalDocs", "")
	if cl == nil {
		return nil, nil
	}

	docs := &openapi.ExternalDocs{}
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
		case "Description":
			docs.Description = val
			hasContent = true
		case "URL":
			docs.URL = val
			hasContent = true
		}
	}
	if !hasContent {
		return nil, nil
	}
	return docs, nil
}
