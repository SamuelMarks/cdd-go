// Package functions parses and emits Go function declarations to/from OpenAPI Operations.
package functions

import (
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/SamuelMarks/cdd-go/src/parameters"
	"github.com/SamuelMarks/cdd-go/src/requestbodies"
	"github.com/SamuelMarks/cdd-go/src/responses"
	"github.com/dave/dst"
)

// ParseOperation converts a dst.FuncDecl into an OpenAPI Operation.
// Parameter fd is the AST function declaration, and it returns the reconstructed OpenAPI Operation and any error.
func ParseOperation(fd *dst.FuncDecl) (*openapi.Operation, error) {
	if fd == nil {
		return nil, fmt.Errorf("FuncDecl is nil")
	}

	op := &openapi.Operation{
		OperationID: fd.Name.Name,
	}

	if fd.Type != nil {
		if fd.Type.Params != nil {
			for _, f := range fd.Type.Params.List {
				if len(f.Names) > 0 && f.Names[0].Name == "body" {
					if rb := requestbodies.Parse(f); rb != nil {
						op.RequestBody = rb
					}
				} else if p := parameters.Parse(f); p != nil {
					op.Parameters = append(op.Parameters, *p)
				}
			}
		}
		if fd.Type.Results != nil && len(fd.Type.Results.List) > 1 {
			op.Responses = responses.Parse(fd.Type.Results)
		}
	}

	if len(fd.Decs.Start) > 0 {
		var cleanDecs dst.Decorations
		for _, dec := range fd.Decs.Start {
			line := strings.TrimSpace(strings.TrimPrefix(dec, "//"))
			if strings.HasPrefix(line, "Path:") || strings.HasPrefix(line, "METHOD:") {
				continue
			}
			cleanDecs = append(cleanDecs, dec)
		}
		if len(cleanDecs) > 0 {
			doc := docstrings.Parse(cleanDecs)
			lines := strings.SplitN(doc, "\n", 2)
			op.Summary = strings.TrimSpace(lines[0])
			if len(lines) > 1 {
				op.Description = strings.TrimSpace(lines[1])
			}
		}
	}

	return op, nil
}
