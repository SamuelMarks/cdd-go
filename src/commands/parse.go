// Package commands provides parsing and emitting of Cobra CLI commands to and from OpenAPI Operations.
package commands

import (
	"strings"

	"github.com/SamuelMarks/cdd-go/src/astutil"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Operation from a Cobra command AST representation.
// Parameter decl is the generic AST declaration, and it returns the HTTP method, URL path, and parsed Operation.
func Parse(decl *dst.GenDecl) (string, string, *openapi.Operation) {
	rawName, cl := astutil.FindSuffixValueCompositeLit(decl, "Cmd")
	if cl == nil {
		return "", "", nil
	}

	var method, path string
	for _, doc := range decl.Decs.Start {
		d := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
		if strings.HasPrefix(d, "Method:") {
			method = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(d, "Method:")))
		} else if strings.HasPrefix(d, "Path:") {
			path = strings.TrimSpace(strings.TrimPrefix(d, "Path:"))
		}
	}

	op := &openapi.Operation{}
	hasContent := false

	opID := strings.TrimSuffix(rawName, "Cmd")
	if opID != "" {
		op.OperationID = strings.ToLower(opID[:1]) + opID[1:]
	}

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
		case "Short":
			op.Summary = val
			hasContent = true
		case "Long":
			op.Description = val
			hasContent = true
		case "Use":
			hasContent = true
		}
	}

	if hasContent || method != "" || path != "" {
		return method, path, op
	}
	return "", "", nil
}
