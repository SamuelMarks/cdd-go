package commands

import (
	"go/token"
	"strconv"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Parse extracts an OpenAPI Operation from a Cobra command AST representation.
func Parse(decl *dst.GenDecl) (string, string, *openapi.Operation) {
	if decl == nil || decl.Tok != token.VAR {
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

	for _, spec := range decl.Specs {
		if vs, ok := spec.(*dst.ValueSpec); ok {
			if len(vs.Names) > 0 && strings.HasSuffix(vs.Names[0].Name, "Cmd") && len(vs.Values) > 0 {
				name := vs.Names[0].Name

				var cl *dst.CompositeLit
				if unary, ok := vs.Values[0].(*dst.UnaryExpr); ok && unary.Op == token.AND {
					cl, _ = unary.X.(*dst.CompositeLit)
				}

				if cl != nil {
					op := &openapi.Operation{}
					hasContent := false

					op.OperationID = strings.TrimSuffix(name, "Cmd")
					if op.OperationID != "" {
						op.OperationID = strings.ToLower(op.OperationID[:1]) + op.OperationID[1:]
					}

					for _, elt := range cl.Elts {
						if kv, ok := elt.(*dst.KeyValueExpr); ok {
							if keyIdent, ok := kv.Key.(*dst.Ident); ok {
								if valLit, ok := kv.Value.(*dst.BasicLit); ok && valLit.Kind == token.STRING {
									val, _ := strconv.Unquote(valLit.Value)
									switch keyIdent.Name {
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
							}
						}
					}

					if hasContent || method != "" || path != "" {
						return method, path, op
					}
				}
			}
		}
	}

	return "", "", nil
}
