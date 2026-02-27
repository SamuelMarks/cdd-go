package functions

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// ParseOperation converts a dst.FuncDecl into an OpenAPI Operation.
func ParseOperation(fd *dst.FuncDecl) (*openapi.Operation, error) {
	if fd == nil {
		return nil, fmt.Errorf("FuncDecl is nil")
	}

	op := &openapi.Operation{
		OperationID: fd.Name.Name,
	}

	if len(fd.Decs.Start) > 0 {
		doc := docstrings.Parse(fd.Decs.Start)
		lines := strings.SplitN(doc, "\n", 2)
		op.Summary = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			op.Description = strings.TrimSpace(lines[1])
		}
	}

	return op, nil
}
