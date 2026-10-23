// Package mocks provides mechanisms to parse and emit mock data from/to OpenAPI Example Objects.
package mocks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SamuelMarks/cdd-go/src/docstrings"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

// ParseExample converts a dst.ValueSpec to an OpenAPI Example.
func ParseExample(vs *dst.ValueSpec) (*openapi.Example, error) {
	if vs == nil {
		return nil, fmt.Errorf("ValueSpec is nil")
	}

	ex := &openapi.Example{}

	if len(vs.Decs.Start) > 0 {
		doc := docstrings.Parse(vs.Decs.Start)
		rawLines := strings.Split(doc, "\n")
		var textLines []string
		for _, line := range rawLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "@dataValue:") {
				ex.DataValue = json.RawMessage(strings.TrimSpace(strings.TrimPrefix(trimmed, "@dataValue:")))
			} else if strings.HasPrefix(trimmed, "@serializedValue:") {
				ex.SerializedValue = strings.TrimSpace(strings.TrimPrefix(trimmed, "@serializedValue:"))
			} else {
				textLines = append(textLines, trimmed)
			}
		}
		if len(textLines) > 0 {
			ex.Summary = textLines[0]
		}
		if len(textLines) > 1 {
			ex.Description = strings.Join(textLines[1:], "\n")
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
