// Package tags provides parsing and emitting of OpenAPI Tag slices to and from Go AST declarations.
package tags

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Tags slice from an AST representation.
// Parameter decl is the generic AST declaration, and it returns the parsed Tag slice and error.
func Parse(decl *dst.GenDecl) ([]openapi.Tag, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}
	_, cl := astutil.FindValueCompositeLit(decl, "Tags", "")
	if cl == nil {
		return nil, nil
	}

	var tags []openapi.Tag
	for _, elt := range cl.Elts {
		tagLit, ok := elt.(*dst.CompositeLit)
		if !ok {
			continue
		}
		tags = append(tags, parseTagElement(tagLit))
	}
	return tags, nil
}

// parseTagElement extracts a single OpenAPI Tag from its CompositeLit element.
// Parameter tagLit is the tag AST composite literal, and it returns the parsed Tag.
func parseTagElement(tagLit *dst.CompositeLit) openapi.Tag {
	tag := openapi.Tag{}
	for _, tagElt := range tagLit.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(tagElt)
		if !ok {
			continue
		}
		val, isStr := astutil.StringLit(valExpr)
		if !isStr {
			continue
		}
		switch key {
		case "Name":
			tag.Name = val
		case "Description":
			tag.Description = val
		}
	}
	return tag
}
