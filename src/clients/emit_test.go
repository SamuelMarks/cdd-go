package clients

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitClientInterface(t *testing.T) {
	pathItem := &openapi.PathItem{
		Summary: "User endpoints",
		Get: &openapi.Operation{
			OperationID: "getUser",
			Summary:     "Get a user by ID",
		},
		Post: &openapi.Operation{
			Summary: "Create a user",
		},
		Put: &openapi.Operation{
			Summary: "Update a user",
		},
		Delete: &openapi.Operation{
			Summary: "Delete a user",
		},
	}

	decl, err := EmitClientInterface("/users/{id}", pathItem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := &dst.File{
		Name:  dst.NewIdent("clients"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err = restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "type // User endpoints") && !strings.Contains(out, "type User endpoints") {
		if !strings.Contains(out, "ClientUsersId interface") {
			t.Errorf("expected interface ClientUsersId, got %s", out)
		}
	}
	if !strings.Contains(out, "// Get a user by ID") {
		t.Errorf("expected Get summary")
	}
	if !strings.Contains(out, "GetUser(req *http.Request) (*http.Response, error)") {
		t.Errorf("expected GetUser method, got %s", out)
	}
	if !strings.Contains(out, "Post(req *http.Request) (*http.Response, error)") {
		t.Errorf("expected Post method, got %s", out)
	}
	if !strings.Contains(out, "Put(req *http.Request) (*http.Response, error)") {
		t.Errorf("expected Put method, got %s", out)
	}
	if !strings.Contains(out, "Delete(req *http.Request) (*http.Response, error)") {
		t.Errorf("expected Delete method, got %s", out)
	}
}

func TestEmitClientInterfaceNil(t *testing.T) {
	_, err := EmitClientInterface("/", nil)
	if err == nil {
		t.Errorf("expected error for nil PathItem")
	}
}

func TestToCamelCase(t *testing.T) {
	if toCamelCase("/") != "Root" {
		t.Errorf("expected Root")
	}
	if toCamelCase("/users") != "Users" {
		t.Errorf("expected Users")
	}
	if toCamelCase("/users/{id}") != "UsersId" {
		t.Errorf("expected UsersId")
	}
}

func TestEmitMethodSignatureEmptyOp(t *testing.T) {
	op := &openapi.Operation{}
	f := emitMethodSignature("Put", op)
	if f.Names[0].Name != "Put" {
		t.Errorf("expected Put")
	}
}
