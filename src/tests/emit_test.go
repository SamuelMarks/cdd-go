package tests

import (
	"strings"
	"testing"

	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitTest(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "getUser",
		Summary:     "Get User",
	}

	fd, err := EmitTest("/users/{id}", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fd.Name.Name != "TestGetUser" {
		t.Errorf("expected TestGetUser, got %s", fd.Name.Name)
	}

	if len(fd.Decs.Start) != 1 {
		t.Errorf("expected 1 line doc")
	} else if !strings.Contains(fd.Decs.Start[0], "tests the Get User operation") {
		t.Errorf("expected summary doc")
	}
}

func TestEmitTestNoOpID(t *testing.T) {
	op := &openapi.Operation{}
	fd, err := EmitTest("/users/{id}", "get", op)
	if err != nil {
		t.Fatal(err)
	}
	if fd.Name.Name != "TestGetUsersId" {
		t.Errorf("expected TestGetUsersId, got %s", fd.Name.Name)
	}
}

func TestEmitTestNil(t *testing.T) {
	_, err := EmitTest("/", "get", nil)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestToCamelCase(t *testing.T) {
	if toCamelCase("/") != "Root" {
		t.Errorf("expected Root")
	}
}
