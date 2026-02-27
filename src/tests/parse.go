package tests

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/openapi"
)

// ParseTest converts a dst.FuncDecl into an OpenAPI Operation representation.
func ParseTest(fd *dst.FuncDecl) (*openapi.Operation, error) {
	if fd == nil {
		return nil, fmt.Errorf("FuncDecl is nil")
	}

	name := fd.Name.Name
	if !strings.HasPrefix(name, "Test") {
		return nil, fmt.Errorf("function is not a test")
	}

	opID := strings.TrimPrefix(name, "Test")
	if opID != "" {
		opID = strings.ToLower(opID[:1]) + opID[1:]
	}

	op := &openapi.Operation{
		OperationID: opID,
	}

	if len(fd.Decs.Start) > 0 {
		desc := ""
		for _, doc := range fd.Decs.Start {
			line := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
			if strings.HasPrefix(line, fd.Name.Name+" tests the ") {
				line = strings.TrimPrefix(line, fd.Name.Name+" tests the ")
				line = strings.TrimSuffix(line, " operation.")
				op.Summary = line
			} else {
				desc += line + " "
			}
		}
		if op.Summary == "" {
			op.Summary = strings.TrimSpace(desc)
		} else {
			op.Description = strings.TrimSpace(desc)
		}
	}

	return op, nil
}
