package docstrings

import (
	"reflect"
	"testing"
)

func TestEmit(t *testing.T) {
	tests := []struct {
		desc     string
		expected []string
	}{
		{"", nil},
		{"simple string", []string{"// simple string"}},
		{"multi\nline\nstring", []string{"// multi", "// line", "// string"}},
	}

	for _, tt := range tests {
		result := Emit(tt.desc)
		if len(result) == 0 && len(tt.expected) == 0 {
			continue
		}
		var strResult []string
		for _, r := range result {
			strResult = append(strResult, r)
		}
		if !reflect.DeepEqual(strResult, tt.expected) {
			t.Errorf("Emit(%q) = %v; want %v", tt.desc, strResult, tt.expected)
		}
	}
}
