// Package securityschemes provides parsing and emitting of OpenAPI SecurityScheme objects to and from Go AST declarations.
package securityschemes

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/oauthflows"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI SecurityScheme object from a struct representation.
// Parameter decl is the generic AST declaration, and it returns the scheme name and parsed SecurityScheme.
func Parse(decl *dst.GenDecl) (string, *openapi.SecurityScheme) {
	if decl == nil || decl.Tok != token.VAR {
		return "", nil
	}
	rawName, cl := astutil.FindValueCompositeLit(decl, "", "SecurityScheme")
	if cl == nil {
		return "", nil
	}

	name := astutil.LowerCamelTrimPrefix(rawName, "SecurityScheme")
	scheme := &openapi.SecurityScheme{}
	hasContent := false

	for _, elt := range cl.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(elt)
		if !ok {
			continue
		}
		if val, isStr := astutil.StringLit(valExpr); isStr {
			if applySecuritySchemeField(scheme, key, val) {
				hasContent = true
			}
			continue
		}
		if clLit, ok := valExpr.(*dst.CompositeLit); ok && key == "Flows" {
			if flows := oauthflows.Parse(clLit); flows != nil {
				scheme.Flows = flows
				hasContent = true
			}
		}
	}

	if !hasContent {
		return "", nil
	}
	return name, scheme
}

// applySecuritySchemeField populates a string field on scheme by field name.
// Parameter scheme is the target SecurityScheme, key is the AST struct field name, and val is the string value.
func applySecuritySchemeField(scheme *openapi.SecurityScheme, key, val string) bool {
	switch key {
	case "Type":
		scheme.Type = val
	case "Description":
		scheme.Description = val
	case "Name":
		scheme.Name = val
	case "In":
		scheme.In = val
	case "Scheme":
		scheme.Scheme = val
	case "BearerFormat":
		scheme.BearerFormat = val
	case "OpenIDConnectURL":
		scheme.OpenIDConnectURL = val
	case "OAuth2MetadataURL":
		scheme.OAuth2MetadataURL = val
	default:
		return false
	}
	return true
}
