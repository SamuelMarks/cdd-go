package tests

import (
	"encoding/json"
	"strings"
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
	_, err := EmitTest("/items/findByStatus", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSynthesizeParamAndJSONPayload(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"Item": {
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":        {Type: "integer"},
				"active":    {Type: "boolean"},
				"tags":      {Type: "array", Items: &openapi.Schema{Type: "string"}},
				"meta":      {Type: "object"},
				"createdAt": {Type: "string", Format: "date-time"},
				"email":     {Type: "string", Format: "email"},
				"status":    {Type: "string", Enum: []json.RawMessage{json.RawMessage(`"available"`)}},
				"empty":     {},
			},
		},
	}

	if ResolveSchema(nil, schemas) != nil {
		t.Errorf("expected nil for nil schema")
	}

	payload := SynthesizeJSONPayload(&openapi.Schema{Ref: "#/components/schemas/Item"}, false, schemas)
	if !strings.Contains(payload, `"status":"available"`) || !strings.Contains(payload, `"email":"dummy@dummy.com"`) {
		t.Errorf("unexpected synthesized payload: %s", payload)
	}

	arrPayload := SynthesizeJSONPayload(&openapi.Schema{Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/Item"}}, false, schemas)
	if !strings.HasPrefix(arrPayload, "[") {
		t.Errorf("expected array payload, got %s", arrPayload)
	}

	if synthesizePropertyJSON("nilProp", nil) != `"dummy"` {
		t.Errorf("expected dummy for nil property")
	}

	// Test SynthesizeParamValue branches
	pEnum := openapi.Parameter{Schema: &openapi.Schema{Type: "string", Enum: []json.RawMessage{json.RawMessage(`"active"`)}}}
	if SynthesizeParamValue(pEnum, schemas) != "active" {
		t.Errorf("expected active")
	}
	pInt := openapi.Parameter{Schema: &openapi.Schema{Type: "integer"}}
	if SynthesizeParamValue(pInt, schemas) != "1" {
		t.Errorf("expected 1")
	}
	pBool := openapi.Parameter{Schema: &openapi.Schema{Type: "boolean"}}
	if SynthesizeParamValue(pBool, schemas) != "true" {
		t.Errorf("expected true")
	}
	pArrEnum := openapi.Parameter{Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Enum: []json.RawMessage{json.RawMessage(`"enumItem"`)}}}}
	if SynthesizeParamValue(pArrEnum, schemas) != "enumItem" {
		t.Errorf("expected enumItem")
	}
	pArrDefault := openapi.Parameter{Schema: &openapi.Schema{Type: "array"}}
	if SynthesizeParamValue(pArrDefault, schemas) != "item1" {
		t.Errorf("expected item1")
	}
	pItemsEnum := openapi.Parameter{Items: &openapi.Items{Enum: []json.RawMessage{json.RawMessage(`"swaggerEnum"`)}}}
	if SynthesizeParamValue(pItemsEnum, schemas) != "swaggerEnum" {
		t.Errorf("expected swaggerEnum")
	}
	pLegacyBool := openapi.Parameter{Type: "boolean"}
	if SynthesizeParamValue(pLegacyBool, schemas) != "true" {
		t.Errorf("expected true")
	}
}

func TestEmitTestCoverageExtras(t *testing.T) {
	op1 := &openapi.Operation{
		OperationID: "op1",
		Parameters: []openapi.Parameter{
			{Name: "username", In: "path"},
			{Name: "userId", In: "path"},
			{Name: "id2", In: "path", Schema: &openapi.Schema{Type: "string"}},
			{Name: "id3", In: "path", Type: "string"},
		},
	}
	if fd, err := EmitTest("/user/{username}/{userId}/{id2}/{id3}", "get", op1); err != nil || fd == nil {
		t.Errorf("expected non-nil FuncDecl and nil err, got %v, %v", fd, err)
	}

	op2 := &openapi.Operation{
		OperationID: "op2",
		Parameters: []openapi.Parameter{
			{Name: "status", In: "query", Required: true},
			{Name: "limit", In: "query", Required: true, Type: "integer"},
		},
	}
	if fd, err := EmitTest("/test", "get", op2); err != nil || fd == nil {
		t.Errorf("expected non-nil FuncDecl and nil err, got %v, %v", fd, err)
	}

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
		if fd, err := EmitTest("/resource", "post", op); err != nil || fd == nil {
			t.Errorf("expected non-nil FuncDecl and nil err, got %v, %v", fd, err)
		}
	}

	contentTypes := []string{"application/x-www-form-urlencoded", "multipart/form-data", "text/plain"}
	for _, cType := range contentTypes {
		op := &openapi.Operation{
			OperationID: "op",
			Consumes:    []string{cType},
			Parameters: []openapi.Parameter{
				{Name: "body", In: "body", Schema: &openapi.Schema{Type: "object"}},
			},
		}
		if fd, err := EmitTest("/test", "post", op); err != nil || fd == nil {
			t.Errorf("expected non-nil FuncDecl and nil err for %s, got %v, %v", cType, fd, err)
		}
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
	if fd, err := EmitTest("/test", "post", op); err != nil || fd == nil {
		t.Errorf("expected non-nil FuncDecl and nil err, got %v, %v", fd, err)
	}
}
