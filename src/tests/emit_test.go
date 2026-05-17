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

func TestEmitTestCoverageExtras(t *testing.T) {
	// 1. Path param string, username
	op1 := &openapi.Operation{
		OperationID: "op1",
		Parameters: []openapi.Parameter{
			{Name: "username", In: "path"},
			{Name: "id2", In: "path", Schema: &openapi.Schema{Type: "string"}},
			{Name: "id3", In: "path", Type: "string"},
		},
	}
	EmitTest("/user/{username}/{id2}/{id3}", "get", op1)

	// 2. Query param status, tags, integer
	op2 := &openapi.Operation{
		OperationID: "op2",
		Parameters: []openapi.Parameter{
			{Name: "status", In: "query", Required: true},
			{Name: "tags", In: "query", Required: true},
			{Name: "limit", In: "query", Required: true, Type: "integer"},
		},
	}
	EmitTest("/test", "get", op2)

	// 3. Various paths and array combinations for body
	paths := []string{"/pet", "/store/order", "/user", "/other"}
	for _, p := range paths {
		for _, isArray := range []bool{true, false} {
			schemaType := "object"
			if isArray {
				schemaType = "array"
			}
			op := &openapi.Operation{
				OperationID: "op",
				Parameters: []openapi.Parameter{
					{Name: "body", In: "body", Schema: &openapi.Schema{Type: schemaType}},
				},
			}
			EmitTest(p, "post", op)
		}
	}

	// 4. Content types
	contentTypes := []string{"application/x-www-form-urlencoded", "multipart/form-data", "text/plain"}
	for _, cType := range contentTypes {
		op := &openapi.Operation{
			OperationID: "op",
			Consumes:    []string{cType},
			Parameters: []openapi.Parameter{
				{Name: "body", In: "body", Schema: &openapi.Schema{Type: "object"}},
			},
		}
		EmitTest("/test", "post", op)
	}
}

func TestEmitTestCoverageExtrasBreakLoop(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "opBreak",
		Consumes:    []string{"something-else", "application/json"},
		Parameters: []openapi.Parameter{
			{Name: "body", In: "body"},
		},
	}
	EmitTest("/test", "post", op)
}
