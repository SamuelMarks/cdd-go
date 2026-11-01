// Package docstrings provides tools for parsing and emitting Go comments for OpenAPI elements.
package docstrings

import (
	"strings"

	"github.com/dave/dst"
)

// Emit formats a string into a dst.Decorations list.
// Parameter desc is the docstring text, and it returns the formatted AST decorations.
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
