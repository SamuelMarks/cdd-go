package routes

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitHandlerInterface(t *testing.T) {
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

	decl, err := EmitHandlerInterface("/users/{id}", pathItem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := &dst.File{
		Name:  dst.NewIdent("routes"),
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
		if !strings.Contains(out, "HandlerUsersId interface") {
			t.Errorf("expected interface HandlerUsersId, got %s", out)
		}
	}
	if !strings.Contains(out, "// Get a user by ID") {
		t.Errorf("expected Get summary")
	}
	if !strings.Contains(out, "GetUser(c *gin.Context)") {
		t.Errorf("expected GetUser method, got %s", out)
	}
	if !strings.Contains(out, "Post(c *gin.Context)") {
		t.Errorf("expected Post method, got %s", out)
	}
	if !strings.Contains(out, "Put(c *gin.Context)") {
		t.Errorf("expected Put method, got %s", out)
	}
	if !strings.Contains(out, "Delete(c *gin.Context)") {
		t.Errorf("expected Delete method, got %s", out)
	}
}

func TestEmitHandlerInterfaceNil(t *testing.T) {
	_, err := EmitHandlerInterface("/", nil)
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

func TestEmitHandlerInterfaceExtraVerbs(t *testing.T) {
	pathItem := &openapi.PathItem{
		Patch: &openapi.Operation{
			OperationID: "patchUser",
		},
		Options: &openapi.Operation{},
		Head:    &openapi.Operation{},
		Trace:   &openapi.Operation{},
	}
	decl, err := EmitHandlerInterface("/extra", pathItem)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	ts := decl.Specs[0].(*dst.TypeSpec)
	iface := ts.Type.(*dst.InterfaceType)
	if len(iface.Methods.List) != 4 {
		t.Errorf("expected 4 methods, got %d", len(iface.Methods.List))
	}
}
