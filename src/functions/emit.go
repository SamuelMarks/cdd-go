// Package functions parses and emits Go function declarations to/from OpenAPI Operations.
package functions

import (
	"fmt"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/naming"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/parameters"
	"github.com/SamuelMarks/cdd-go/src/requestbodies"
	"github.com/SamuelMarks/cdd-go/src/responses"
	"github.com/dave/dst"
)

// EmitOperation generates a dst.FuncDecl from an OpenAPI Operation.
// Parameter op is the OpenAPI Operation, and it returns the AST function declaration and any error.
func EmitOperation(op *openapi.Operation) (*dst.FuncDecl, error) {
	if op == nil {
		return nil, fmt.Errorf("Operation is nil")
	}

	name := "GeneratedOperation"
	if op.OperationID != "" {
		name = naming.ToPascalCase(op.OperationID)
	}

	paramFields := &dst.FieldList{}
	for _, p := range op.Parameters {
		if f := parameters.Emit(p); f != nil {
			paramFields.List = append(paramFields.List, f)
		}
	}
	if op.RequestBody != nil {
		if rbField := requestbodies.Emit(op.RequestBody); rbField != nil {
			paramFields.List = append(paramFields.List, rbField)
		}
	}

	resultFields := &dst.FieldList{
		List: []*dst.Field{
			{Type: dst.NewIdent("error")},
		},
	}
	if len(op.Responses) > 0 {
		resultFields.List = nil
		for _, expr := range responses.Emit(op.Responses) {
			resultFields.List = append(resultFields.List, &dst.Field{Type: expr})
		}
	}

	fd := &dst.FuncDecl{
		Name: dst.NewIdent(name),
		Type: &dst.FuncType{
			Params:  paramFields,
			Results: resultFields,
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
