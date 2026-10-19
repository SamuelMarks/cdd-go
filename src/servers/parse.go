// Package servers provides parsing and emitting of OpenAPI Server slices to and from Go AST declarations.
package servers

import (
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Servers slice from an AST representation.
// Parameter decl is the generic AST declaration, and it returns the parsed Server slice and error.
func Parse(decl *dst.GenDecl) ([]openapi.Server, error) {
	if decl == nil || decl.Tok != token.VAR {
		return nil, nil
	}
	_, cl := astutil.FindValueCompositeLit(decl, "Servers", "")
	if cl == nil {
		return nil, nil
	}

	var servers []openapi.Server
	for _, elt := range cl.Elts {
		srvLit, ok := elt.(*dst.CompositeLit)
		if !ok {
			continue
		}
		servers = append(servers, parseServerElement(srvLit))
	}
	return servers, nil
}

// parseServerElement extracts a single OpenAPI Server from its CompositeLit element.
// Parameter srvLit is the server AST composite literal, and it returns the parsed Server.
func parseServerElement(srvLit *dst.CompositeLit) openapi.Server {
	srv := openapi.Server{}
	for _, srvElt := range srvLit.Elts {
		key, valExpr, ok := astutil.IdentKeyValue(srvElt)
		if !ok {
			continue
		}
		val, isStr := astutil.StringLit(valExpr)
		if !isStr {
			continue
		}
		switch key {
		case "URL":
			srv.URL = val
		case "Description":
			srv.Description = val
		}
	}
	return srv
}
