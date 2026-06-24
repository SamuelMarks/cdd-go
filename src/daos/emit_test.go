package daos

import (
	"bytes"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitDAOInterface(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
		},
	}

	decl, err := EmitDAOInterface("User", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: []dst.Decl{decl},
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("UserDAO")) {
		t.Errorf("missing UserDAO in code")
	}
	if !bytes.Contains([]byte(code), []byte("Create(ctx context.Context, item *User) error")) {
		t.Errorf("missing Create method in code")
	}
}

func TestEmitStubDAO(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
		},
	}

	decl, methods, err := EmitStubDAO("User", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}
	if len(methods) != 5 {
		t.Fatalf("expected 5 methods, got %d", len(methods))
	}
}

func TestEmitConcreteDAO(t *testing.T) {
	schema := &openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"id":   {Type: "string"},
			"name": {Type: "string"},
		},
	}

	decl, methods, err := EmitConcreteDAO("User", schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}
	if len(methods) != 5 {
		t.Fatalf("expected 5 methods, got %d", len(methods))
	}
}

func TestEmitDAONilSchema(t *testing.T) {
	_, err := EmitDAOInterface("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	_, _, err = EmitStubDAO("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	_, _, err = EmitConcreteDAO("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
}
