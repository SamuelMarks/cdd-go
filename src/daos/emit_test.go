package daos

import (
	"bytes"
	"strings"
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
	if !bytes.Contains([]byte(code), []byte("Create(ctx context.Context, item *models.User) error")) {
		t.Errorf("missing Create method with *models.User in code: %s", code)
	}
	if !bytes.Contains([]byte(code), []byte("List(ctx context.Context) ([]*models.User, error)")) {
		t.Errorf("missing List method with []*models.User in code: %s", code)
	}

	bundle, err := EmitSchemaDAOs("User", schema)
	if err != nil || bundle == nil || bundle.InterfaceDecl == nil || bundle.StubDecl == nil || bundle.ConcreteDecl == nil {
		t.Fatalf("expected valid SchemaDAOBundle, got %+v, %v", bundle, err)
	}
	if _, err := EmitSchemaDAOs("", schema); err == nil {
		t.Errorf("expected error from EmitSchemaDAOs with empty name")
	}

	errFile := EmitModelsErrorsFile()
	var errBuf bytes.Buffer
	if err := restorer.Fprint(&errBuf, errFile); err != nil {
		t.Fatalf("failed to print models errors file: %v", err)
	}
	if !strings.Contains(errBuf.String(), "func NewNotImplementedError() error") {
		t.Errorf("expected NewNotImplementedError in models/errors.go, got: %s", errBuf.String())
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

	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: append([]dst.Decl{decl}, methods...),
	}
	if err := decorator.NewRestorer().Fprint(&buf, file); err != nil {
		t.Fatalf("failed to render stub DAO: %v", err)
	}
	code := buf.String()
	if !strings.Contains(code, "UserStubDAO") || !strings.Contains(code, "models.NewNotImplementedError()") {
		t.Errorf("expected UserStubDAO and models.NewNotImplementedError() in rendered code")
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

	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: append([]dst.Decl{decl}, methods...),
	}
	if err := decorator.NewRestorer().Fprint(&buf, file); err != nil {
		t.Fatalf("failed to render concrete DAO: %v", err)
	}
	code := buf.String()
	for _, expected := range []string{"UserGormDAO", "WithContext(ctx)", ".Create(", ".First(", ".Save(", ".Delete(", ".Find("} {
		if !strings.Contains(code, expected) {
			t.Errorf("expected %q in rendered ConcreteDAO code", expected)
		}
	}
}

func TestEmitDAONilSchema(t *testing.T) {
	_, err := EmitDAOInterface("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	_, err = EmitDAOInterface("", &openapi.Schema{Type: "object"})
	if err == nil {
		t.Error("expected error for empty schema name")
	}
	_, err = EmitDAOInterface("User", &openapi.Schema{Type: "invalid-type"})
	if err == nil {
		t.Error("expected error for invalid schema type")
	}
	_, _, err = EmitStubDAO("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	_, _, err = EmitStubDAO("", &openapi.Schema{Type: "object"})
	if err == nil {
		t.Error("expected error for empty schema name")
	}
	_, _, err = EmitStubDAO("User", &openapi.Schema{Type: "invalid-type"})
	if err == nil {
		t.Error("expected error for invalid schema type")
	}
	_, _, err = EmitConcreteDAO("User", nil)
	if err == nil {
		t.Error("expected error for nil schema")
	}
	_, _, err = EmitConcreteDAO("", &openapi.Schema{Type: "object"})
	if err == nil {
		t.Error("expected error for empty schema name")
	}
	_, _, err = EmitConcreteDAO("User", &openapi.Schema{Type: "invalid-type"})
	if err == nil {
		t.Error("expected error for invalid schema type")
	}
}
