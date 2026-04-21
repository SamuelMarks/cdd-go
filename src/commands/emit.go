package commands

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// Emit formats an OpenAPI Operation object into a Cobra command AST representation.
func Emit(path, method string, op *openapi.Operation) *dst.GenDecl {
	if op == nil {
		return nil
	}

	opID := op.OperationID
	if opID == "" {
		opID = strings.ToLower(method) + toPascalCase(strings.ReplaceAll(path, "/", "_"))
		opID = strings.ReplaceAll(opID, "{", "")
		opID = strings.ReplaceAll(opID, "}", "")
	}

	useName := strings.ToLower(opID)

	cl := &dst.CompositeLit{
		Type: &dst.SelectorExpr{
			X:   dst.NewIdent("cobra"),
			Sel: dst.NewIdent("Command"),
		},
		Elts: []dst.Expr{
			&dst.KeyValueExpr{
				Key:   dst.NewIdent("Use"),
				Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", useName)},
			},
		},
	}

	if op.Summary != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Short"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", op.Summary)},
		})
	}

	if op.Description != "" {
		cl.Elts = append(cl.Elts, &dst.KeyValueExpr{
			Key:   dst.NewIdent("Long"),
			Value: &dst.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", op.Description)},
		})
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names: []*dst.Ident{dst.NewIdent(toPascalCase(opID) + "Cmd")},
				Values: []dst.Expr{
					&dst.UnaryExpr{
						Op: token.AND,
						X:  cl,
					},
				},
			},
		},
	}

	decl.Decs.Start.Append(fmt.Sprintf("// Method: %s", strings.ToUpper(method)))
	decl.Decs.Start.Append(fmt.Sprintf("// Path: %s", path))

	return decl
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	var res string
	for _, p := range parts {
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return res
}
