package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitCommand(t *testing.T) {
	op := &openapi.Operation{
		OperationID: "getUser",
		Summary:     "Get user",
		Description: "Gets a user by ID",
	}

	decl := Emit("/users/{id}", "get", op)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("commands"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "\n", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "// Method: GET") {
		t.Errorf("missing method comment")
	}
	if !strings.Contains(out, "// Path: /users/{id}") {
		t.Errorf("missing path comment")
	}
	if !strings.Contains(out, "var GetUserCmd = &cobra.Command{") {
		t.Errorf("missing command var, got %s", out)
	}
	if !strings.Contains(out, `Use: "getuser"`) {
		t.Errorf("missing use")
	}
	if !strings.Contains(out, `Short: "Get user"`) {
		t.Errorf("missing short")
	}
	if !strings.Contains(out, `Long: "Gets a user by ID"`) {
		t.Errorf("missing long")
	}
}

func TestEmitCommandNoOpID(t *testing.T) {
	op := &openapi.Operation{}
	decl := Emit("/users", "post", op)
	if decl == nil {
		t.Fatalf("expected decl")
	}
	vs := decl.Specs[0].(*dst.ValueSpec)
	if vs.Names[0].Name != "PostUsersCmd" {
		t.Errorf("expected PostUsersCmd, got %s", vs.Names[0].Name)
	}
}

func TestEmitCommandNil(t *testing.T) {
	if Emit("/", "get", nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestToPascalCaseEmpty(t *testing.T) {
	if toPascalCase("") != "" {
		t.Errorf("expected empty")
	}
	if toPascalCase("Test") != "Test" {
		t.Errorf("expected Test")
	}
}
