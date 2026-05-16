package tests

import (
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

func TestEmitTestNil(t *testing.T) {
	_, err := EmitTest("/test", "get", nil)
	if err == nil {
		t.Errorf("expected error for nil operation")
	}
}

func TestEmitTestNoOpID(t *testing.T) {
	op := &openapi.Operation{
		Summary: "Test op",
	}
	fd, err := EmitTest("/test/path", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fd.Name.Name != "TestGetTestPath" {
		t.Errorf("expected TestGetTestPath, got %s", fd.Name.Name)
	}
}

func TestEmitTestWithBodyAndPathParams(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "testOp",
		Parameters: []openapi.Parameter{
			{Name: "id", In: "path"},
			{Name: "body", In: "body", Schema: &openapi.Schema{Type: "array"}},
		},
	}
	_, err := EmitTest("/test/{id}", "post", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmitTestRequestBodyArray(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "testOp2",
		RequestBody: &openapi.RequestBody{
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{Type: "array"},
				},
			},
		},
	}
	_, err := EmitTest("/test2", "post", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToCamelCase(t *testing.T) {
	if toCamelCase("/") != "Root" {
		t.Errorf("expected Root")
	}
	if toCamelCase("/test/{id}") != "TestId" {
		t.Errorf("expected TestId")
	}
}

func TestEmitTestBodyParamGet(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "testOp",
		Parameters: []openapi.Parameter{
			{Name: "body", In: "body"},
		},
	}
	_, err := EmitTest("/test", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmitTestFindByStatusGet(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "findByStatus",
	}
	_, err := EmitTest("/pet/findByStatus", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
