package daos

import (
	"bytes"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitFactoryInterface(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"User": {Type: "object"},
		"Post": {Type: "object"},
	}

	decl := EmitFactoryInterface(schemas)
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
	if !bytes.Contains([]byte(code), []byte("Factory")) {
		t.Errorf("missing Factory interface in code")
	}
	if !bytes.Contains([]byte(code), []byte("User() UserDAO")) {
		t.Errorf("missing User() method in code")
	}
}

func TestEmitConcreteFactory(t *testing.T) {
	schemas := map[string]openapi.Schema{
		"User": {Type: "object"},
	}

	decl, methods := EmitConcreteFactory(schemas)
	if decl == nil {
		t.Fatal("expected non-nil decl")
	}
	if len(methods) != 2 { // NewDAOFactory + User()
		t.Fatalf("expected 2 methods, got %d", len(methods))
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: append([]dst.Decl{decl}, methods...),
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("DAOFactory")) {
		t.Errorf("missing DAOFactory struct in code")
	}
	if !bytes.Contains([]byte(code), []byte("NewDAOFactory")) {
		t.Errorf("missing NewDAOFactory func in code")
	}
}
