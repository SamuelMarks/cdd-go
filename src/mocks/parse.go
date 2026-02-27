package mocks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/samuel/cdd-go/src/docstrings"
	"github.com/samuel/cdd-go/src/openapi"
)

// ParseExample converts a dst.ValueSpec to an OpenAPI Example.
func ParseExample(vs *dst.ValueSpec) (*openapi.Example, error) {
	if vs == nil {
		return nil, fmt.Errorf("ValueSpec is nil")
	}

	ex := &openapi.Example{}

	if len(vs.Decs.Start) > 0 {
		doc := docstrings.Parse(vs.Decs.Start)
		lines := strings.SplitN(doc, "\n", 2)
		ex.Summary = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			ex.Description = strings.TrimSpace(lines[1])
		}
	}

	if len(vs.Values) > 0 {
		if bl, ok := vs.Values[0].(*dst.BasicLit); ok {
			val := strings.TrimPrefix(bl.Value, "`")
			val = strings.TrimSuffix(val, "`")
			if val == `""` {
				val = ""
			}
			ex.Value = json.RawMessage(val)
		}
	}

	return ex, nil
}
