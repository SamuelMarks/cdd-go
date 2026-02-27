package docstrings

import (
	"testing"

	"github.com/dave/dst"
)

func TestParse(t *testing.T) {
	tests := []struct {
		decs     dst.Decorations
		expected string
	}{
		{nil, ""},
		{dst.Decorations{"// simple string"}, "simple string"},
		{dst.Decorations{"// multi", "// line", "// string"}, "multi\nline\nstring"},
	}

	for _, tt := range tests {
		result := Parse(tt.decs)
		if result != tt.expected {
			t.Errorf("Parse(%v) = %q; want %q", tt.decs, result, tt.expected)
		}
	}
}
