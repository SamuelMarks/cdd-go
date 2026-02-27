package docstrings

import (
	"strings"

	"github.com/dave/dst"
)

// Parse converts a dst.Decorations list into a single string.
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
