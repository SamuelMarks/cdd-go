// Package mocks provides mechanisms to parse and emit mock data from/to OpenAPI Example Objects.
package mocks

import (
	"fmt"
	"go/token"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// EmitExample generates a dst.GenDecl (variable declaration) from an OpenAPI Example.
func EmitExample(name string, ex *openapi.Example) (*dst.GenDecl, error) {
	if ex == nil {
		return nil, fmt.Errorf("Example is nil")
	}
	if name == "" || !token.IsIdentifier(name) {
		return nil, fmt.Errorf("invalid example identifier name: %q", name)
	}

	valStr := string(ex.Value)
	if valStr == "" && len(ex.DataValue) > 0 {
		valStr = string(ex.DataValue)
	}
	if valStr == "" && ex.SerializedValue != "" {
		valStr = ex.SerializedValue
	}
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

	var docLines []string
	if ex.Summary != "" {
		docLines = append(docLines, ex.Summary)
	}
	if ex.Description != "" {
		docLines = append(docLines, ex.Description)
	}
	if len(ex.DataValue) > 0 {
		docLines = append(docLines, "@dataValue: "+string(ex.DataValue))
	}
	if ex.SerializedValue != "" {
		docLines = append(docLines, "@serializedValue: "+ex.SerializedValue)
	}
	if len(docLines) > 0 {
		desc := docLines[0]
		for _, line := range docLines[1:] {
			desc += "\n" + line
		}
		vs.Decs.Start = docstrings.Emit(desc)
	}

	return &dst.GenDecl{
		Tok:   token.VAR,
		Specs: []dst.Spec{vs},
	}, nil
}
