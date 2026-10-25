package tests

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

// renderFuncDecl renders a dst.FuncDecl into Go source code for behavioral assertions.
func renderFuncDecl(t *testing.T, fd *dst.FuncDecl) string {
	t.Helper()
	if fd == nil {
		t.Fatal("expected non-nil FuncDecl")
	}
	f := &dst.File{
		Name:  dst.NewIdent("tests"),
		Decls: []dst.Decl{fd},
	}
	var buf bytes.Buffer
	if err := decorator.NewRestorer().Fprint(&buf, f); err != nil {
		t.Fatalf("failed to render FuncDecl: %v", err)
	}
	return buf.String()
}

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
	fd, hasBody, err := EmitTestWithConfig("/test/path", "get", op, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hasBody {
		t.Errorf("did not expect body for GET /test/path")
	}
	if fd.Name.Name != "TestGetTestPath" {
		t.Errorf("expected TestGetTestPath, got %s", fd.Name.Name)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, `http.NewRequest("GET", baseURL+"/test/path", nil)`) {
		t.Errorf("expected GET request to /test/path, got: %s", code)
	}
	parsedOp, err := ParseTest(fd)
	if err != nil || parsedOp.Summary != "Test op" {
		t.Errorf("expected deterministic round-trip of Summary 'Test op', got %+v, %v", parsedOp, err)
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
	fd, err := EmitTest("/test/{id}", "post", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, `http.NewRequest("POST", baseURL+"/test/1"`) {
		t.Errorf("expected interpolated path /test/1, got: %s", code)
	}
	if !strings.Contains(code, `\"id\":1`) {
		t.Errorf("expected synthesized JSON array body, got: %s", code)
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
	fd, err := EmitTest("/test2", "post", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, `req.Header.Set("Content-Type", "application/json")`) {
		t.Errorf("expected Content-Type header, got: %s", code)
	}
}

func TestEmitTestBodyParamGet(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "testOp",
		Parameters: []openapi.Parameter{
			{Name: "body", In: "body"},
		},
	}
	fd, err := EmitTest("/test", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, `strings.NewReader`) {
		t.Errorf("expected body reader in GET with body param, got: %s", code)
	}
}

func TestEmitTestFindByStatusGet(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "findByStatus",
	}
	fd, err := EmitTest("/items/findByStatus", "get", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, "func TestFindByStatus(t *testing.T)") {
		t.Errorf("expected TestFindByStatus function, got: %s", code)
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
	fd1, err := EmitTest("/user/{username}/{userId}/{id2}/{id3}", "get", op1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code1 := renderFuncDecl(t, fd1)
	if !strings.Contains(code1, `/user/dummy/1/dummy/dummy`) {
		t.Errorf("expected interpolated path parameters, got: %s", code1)
	}

	op2 := &openapi.Operation{
		OperationID: "op2",
		Parameters: []openapi.Parameter{
			{Name: "status", In: "query", Required: true},
			{Name: "limit", In: "query", Required: true, Type: "integer"},
		},
	}
	fd2, err := EmitTest("/test", "get", op2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code2 := renderFuncDecl(t, fd2)
	if !strings.Contains(code2, `/test?status=dummy&limit=1`) {
		t.Errorf("expected query parameters in URL, got: %s", code2)
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
		fd, err := EmitTest("/resource", "post", op)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = renderFuncDecl(t, fd)
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
		fd, err := EmitTest("/test", "post", op)
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", cType, err)
		}
		code := renderFuncDecl(t, fd)
		if !strings.Contains(code, cType) {
			t.Errorf("expected Content-Type %s in generated code: %s", cType, code)
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
	fd, err := EmitTest("/test", "post", op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	code := renderFuncDecl(t, fd)
	if !strings.Contains(code, `"application/json"`) {
		t.Errorf("expected application/json Content-Type, got: %s", code)
	}
}

func TestEmitTestResponseSchemaValidation(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"Pet": {
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":   {Type: "integer"},
				"name": {Type: "string"},
			},
		},
		"ImplicitObject": {
			Properties: map[string]openapi.Schema{
				"code": {Type: "integer"},
			},
		},
	}

	if resolveSuccessResponseSchema(nil, schemas) != nil {
		t.Errorf("expected nil for nil operation")
	}

	cases := []struct {
		name     string
		op       *openapi.Operation
		expected string
	}{
		{
			name: "object_ref_json",
			op: &openapi.Operation{
				OperationID: "getPet",
				Responses: openapi.Responses{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {Schema: &openapi.Schema{Ref: "#/components/schemas/Pet"}},
						},
					},
				},
			},
			expected: "matchedField",
		},
		{
			name: "array_swagger_schema",
			op: &openapi.Operation{
				OperationID: "listPets",
				Responses: openapi.Responses{
					"200": {
						Schema: &openapi.Schema{Type: "array", Items: &openapi.Schema{Ref: "#/components/schemas/Pet"}},
					},
				},
			},
			expected: "var dummyVal []json.RawMessage",
		},
		{
			name: "string_non_json_content",
			op: &openapi.Operation{
				OperationID: "getStatus",
				Responses: openapi.Responses{
					"default": {
						Content: map[string]openapi.MediaType{
							"text/plain": {Schema: &openapi.Schema{Type: "string"}},
						},
					},
				},
			},
			expected: "var dummyVal string",
		},
		{
			name: "integer_response",
			op: &openapi.Operation{
				OperationID: "getCount",
				Responses: openapi.Responses{
					"200": {Schema: &openapi.Schema{Type: "integer"}},
				},
			},
			expected: "var dummyVal int64",
		},
		{
			name: "number_response",
			op: &openapi.Operation{
				OperationID: "getScore",
				Responses: openapi.Responses{
					"200": {Schema: &openapi.Schema{Type: "number"}},
				},
			},
			expected: "var dummyVal float64",
		},
		{
			name: "boolean_response",
			op: &openapi.Operation{
				OperationID: "getActive",
				Responses: openapi.Responses{
					"200": {Schema: &openapi.Schema{Type: "boolean"}},
				},
			},
			expected: "var dummyVal bool",
		},
		{
			name: "implicit_object_properties",
			op: &openapi.Operation{
				OperationID: "getImplicit",
				Responses: openapi.Responses{
					"200": {Schema: &openapi.Schema{Ref: "#/components/schemas/ImplicitObject"}},
				},
			},
			expected: "matchedField",
		},
		{
			name: "error_only_response",
			op: &openapi.Operation{
				OperationID: "getErrOnly",
				Responses: openapi.Responses{
					"404": {Description: "not found"},
				},
			},
			expected: "var dummyVal json.RawMessage",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fd, err := EmitTestWithSchemas("/test", "get", tc.op, schemas)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			code := renderFuncDecl(t, fd)
			if !strings.Contains(code, tc.expected) {
				t.Errorf("expected %q in generated code, got:\n%s", tc.expected, code)
			}
			if !strings.Contains(code, `"__sabotage__"`) {
				t.Errorf("expected __sabotage__ check in generated code, got:\n%s", code)
			}
		})
	}
}
