// Package docstrings provides tools for parsing and emitting Go comments for OpenAPI elements.
package docstrings

import (
	"strings"

	"github.com/dave/dst"
)

// Parse converts a dst.Decorations list into a single string.
// Parameter decs is the slice of comment decorations, and it returns the newline-joined docstring.
func Parse(decs dst.Decorations) string {
	if len(decs) == 0 {
		return ""
	}
	var res []string
	for _, dec := range decs {
		res = append(res, strings.TrimSpace(strings.TrimPrefix(dec, "//")))
	}
	return strings.Join(res, "\n")
}

// ParseSummaryAndTag extracts a space-joined summary string and an optional tagged value (e.g. "METHOD:" or "Path:") from a dst.Decorations slice.
// Parameter decs is the slice of comment decorations, tagPrefix is the tag prefix to extract (such as "METHOD:"), and it returns the summary and extracted tag value.
func ParseSummaryAndTag(decs dst.Decorations, tagPrefix string) (string, string) {
	if len(decs) == 0 {
		return "", ""
	}
	var descParts []string
	var tagVal string
	for _, doc := range decs {
		line := strings.TrimSpace(strings.TrimPrefix(doc, "//"))
		if tagPrefix != "" && strings.HasPrefix(line, tagPrefix) {
			tagVal = strings.TrimSpace(strings.TrimPrefix(line, tagPrefix))
			continue
		}
		if line != "" {
			descParts = append(descParts, line)
		}
	}
	return strings.Join(descParts, " "), tagVal
}
