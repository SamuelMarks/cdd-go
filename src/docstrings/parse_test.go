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

func TestParseSummaryAndTag(t *testing.T) {
	if s, tag := ParseSummaryAndTag(nil, "METHOD:"); s != "" || tag != "" {
		t.Errorf("expected empty for nil decs, got %q, %q", s, tag)
	}
	decs := dst.Decorations{
		"// METHOD: POST",
		"// Create a user",
		"// in the database",
	}
	s, tag := ParseSummaryAndTag(decs, "METHOD:")
	if s != "Create a user in the database" || tag != "POST" {
		t.Errorf("unexpected ParseSummaryAndTag result: %q, %q", s, tag)
	}
}
