// Package docstrings provides tools for parsing and emitting Go comments for OpenAPI elements.
package docstrings

import (
	"strings"

	"github.com/dave/dst"
)

// Emit formats a string into a dst.Decorations list.
func Emit(desc string) dst.Decorations {
	if desc == "" {
		return nil
	}
	var decs dst.Decorations
	lines := strings.Split(desc, "\n")
	for _, line := range lines {
		decs = append(decs, "// "+line)
	}
	return decs
}
