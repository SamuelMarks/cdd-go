package daos

import (
	"bytes"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitFactoryTests(t *testing.T) {
	schemas := map[string]openapi.Schema{"User": {Type: "object"}}
	_, decls := EmitFactoryTests(schemas)
	if len(decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(decls))
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: decls,
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("TestDAOFactory")) {
		t.Errorf("missing TestDAOFactory in code")
	}
}

func TestEmitConcreteDAOTests(t *testing.T) {
	schemas := map[string]openapi.Schema{"User": {Type: "object"}}
	decls := EmitConcreteDAOTests(schemas)
	if len(decls) != 1 {
		t.Fatalf("expected 1 decl, got %d", len(decls))
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	file := &dst.File{
		Name:  dst.NewIdent("daos"),
		Decls: decls,
	}
	if err := restorer.Fprint(&buf, file); err != nil {
		t.Fatalf("failed to print decl: %v", err)
	}

	code := buf.String()
	if !bytes.Contains([]byte(code), []byte("TestUserConcreteDAO")) {
		t.Errorf("missing TestUserConcreteDAO in code")
	}
}
