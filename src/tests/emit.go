// Package tests provides parsing and emitting for Go test functions for OpenAPI paths/operations.
package tests

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// EmitTest generates a dst.FuncDecl for testing an OpenAPI operation.
func EmitTest(path string, method string, op *openapi.Operation) (*dst.FuncDecl, error) {
	if op == nil {
		return nil, fmt.Errorf("Operation is nil")
	}

	name := "Test"
	if op.OperationID != "" {
		name += strings.ToUpper(op.OperationID[:1]) + op.OperationID[1:]
	} else {
		// Use path and method
		pathCamel := toCamelCase(path)
		name += strings.ToUpper(method[:1]) + method[1:] + pathCamel
	}

	fd := &dst.FuncDecl{
		Name: dst.NewIdent(name),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{
						Names: []*dst.Ident{dst.NewIdent("t")},
						Type: &dst.StarExpr{
							X: &dst.SelectorExpr{
								X:   dst.NewIdent("testing"),
								Sel: dst.NewIdent("T"),
							},
						},
					},
				},
			},
		},
		Body: &dst.BlockStmt{},
	}

	if op.Summary != "" {
		fd.Decs.Start.Append(fmt.Sprintf("// %s tests the %s operation.", name, op.Summary))
	}

	return fd, nil
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "/")
	var res string
	for _, p := range parts {
		p = strings.ReplaceAll(p, "{", "")
		p = strings.ReplaceAll(p, "}", "")
		if p != "" {
			res += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	if res == "" {
		res = "Root"
	}
	return res
}
