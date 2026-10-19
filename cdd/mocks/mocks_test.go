package mocks

import (
	"testing"

	srcmocks "github.com/SamuelMarks/cdd-go/src/mocks"
	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestDefaultEmptyMock(t *testing.T) {
	if DefaultEmptyMock != `""` {
		t.Errorf("expected empty string mock literal, got %q", DefaultEmptyMock)
	}

	decl, err := srcmocks.EmitExample("EmptyEx", &openapi.Example{})
	if err != nil || decl == nil {
		t.Fatalf("expected non-nil decl and nil error, got %v, %v", decl, err)
	}
	vs := decl.Specs[0].(*dst.ValueSpec)
	parsedEx, err := srcmocks.ParseExample(vs)
	if err != nil || string(parsedEx.Value) != "" {
		t.Errorf("expected empty mock value after round-trip, got %q, %v", string(parsedEx.Value), err)
	}
}
