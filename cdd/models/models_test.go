package models

import "testing"

func TestAppError(t *testing.T) {
	err := NewNotImplementedError()
	if err.Error() != "Not implemented" {
		t.Errorf("unexpected message")
	}
}
