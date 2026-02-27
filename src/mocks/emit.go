// Package mocks provides mechanisms to parse and emit mock data from/to OpenAPI Example Objects.
package mocks

import (
	"fmt"
	"go/token"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// EmitExample generates a dst.GenDecl (variable declaration) from an OpenAPI Example.
func EmitExample(name string, ex *openapi.Example) (*dst.GenDecl, error) {
	if ex == nil {
		return nil, fmt.Errorf("Example is nil")
	}

	valStr := string(ex.Value)
	if valStr == "" {
		valStr = `""`
	}

	vs := &dst.ValueSpec{
		Names: []*dst.Ident{dst.NewIdent(name)},
		Type:  dst.NewIdent("string"),
		Values: []dst.Expr{
			&dst.BasicLit{
				Kind:  token.STRING,
				Value: "`" + valStr + "`",
			},
		},
	}

	if ex.Summary != "" || ex.Description != "" {
		desc := ex.Summary
		if ex.Description != "" {
			if desc != "" {
				desc += "\n"
			}
			desc += ex.Description
		}
		vs.Decs.Start = docstrings.Emit(desc)
	}

	return &dst.GenDecl{
		Tok:   token.VAR,
		Specs: []dst.Spec{vs},
	}, nil
}
