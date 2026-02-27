package functions

import (
	"strings"
	"testing"

	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitOperation(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "CreateUser",
		Summary:     "Creates a new user",
		Description: "Detailed description",
	}

	fd, err := EmitOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fd.Name.Name != "CreateUser" {
		t.Errorf("expected CreateUser")
	}

	if len(fd.Decs.Start) != 2 {
		t.Errorf("expected 2 lines of docstring, got %d", len(fd.Decs.Start))
	} else {
		if !strings.Contains(fd.Decs.Start[0], "Creates a new user") {
			t.Errorf("expected summary, got %s", fd.Decs.Start[0])
		}
		if !strings.Contains(fd.Decs.Start[1], "Detailed description") {
			t.Errorf("expected description, got %s", fd.Decs.Start[1])
		}
	}
}

func TestEmitOperationDefaults(t *testing.T) {
	op := &openapi.Operation{}
	fd, err := EmitOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fd.Name.Name != "GeneratedOperation" {
		t.Errorf("expected default name")
	}
	if len(fd.Decs.Start) != 0 {
		t.Errorf("expected 0 lines of docstring")
	}
}

func TestEmitOperationSummaryOnly(t *testing.T) {
	op := &openapi.Operation{Summary: "Sum"}
	fd, err := EmitOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fd.Decs.Start) != 1 {
		t.Errorf("expected 1 line of docstring")
	}
}

func TestEmitOperationDescOnly(t *testing.T) {
	op := &openapi.Operation{Description: "Desc"}
	fd, err := EmitOperation(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fd.Decs.Start) != 1 {
		t.Errorf("expected 1 line of docstring")
	}
}

func TestEmitOperationNil(t *testing.T) {
	_, err := EmitOperation(nil)
	if err == nil {
		t.Errorf("expected error")
	}
}
