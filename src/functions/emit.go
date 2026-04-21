// Package functions parses and emits Go function declarations to/from OpenAPI Operations.
package functions

import (
	"fmt"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitOperation generates a dst.FuncDecl from an OpenAPI Operation.
func EmitOperation(op *openapi.Operation) (*dst.FuncDecl, error) {
	if op == nil {
		return nil, fmt.Errorf("Operation is nil")
	}

	name := "GeneratedOperation"
	if op.OperationID != "" {
		name = op.OperationID
	}

	fd := &dst.FuncDecl{
		Name: dst.NewIdent(name),
		Type: &dst.FuncType{
			Params: &dst.FieldList{},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Body: &dst.BlockStmt{},
	}

	if op.Summary != "" || op.Description != "" {
		desc := op.Summary
		if op.Description != "" {
			if desc != "" {
				desc += "\n"
			}
			desc += op.Description
		}
		fd.Decs.Start = docstrings.Emit(desc)
	}

	return fd, nil
}
