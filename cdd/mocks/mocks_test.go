package mocks

import "testing"

func TestDefaultEmptyMock(t *testing.T) {
	if DefaultEmptyMock != `""` {
		t.Errorf("expected empty string mock literal, got %q", DefaultEmptyMock)
	}
}
