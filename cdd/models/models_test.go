package models

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	err := NewNotImplementedError()
	if err.Error() != "Not implemented" {
		t.Errorf("unexpected message: %q", err.Error())
	}

	var appErr *AppError
	if !errors.As(err, &appErr) || appErr.Kind != ErrKindNotImplemented || appErr.Code != 501 {
		t.Errorf("expected AppError with ErrKindNotImplemented and 501, got %+v", appErr)
	}
	if appErr.Unwrap() != nil {
		t.Errorf("expected nil cause")
	}

	rootCause := errors.New("disk full")
	wrapped := NewDomainError(ErrKindCodegenIO, 500, "failed writing AST", rootCause)
	if !errors.Is(wrapped, rootCause) {
		t.Errorf("expected errors.Is to match rootCause")
	}
	if wrapped.Error() != "failed writing AST: disk full" {
		t.Errorf("unexpected wrapped error string: %q", wrapped.Error())
	}
}
