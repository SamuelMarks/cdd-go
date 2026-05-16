package links

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitLink(t *testing.T) {
	link := &openapi.Link{
		OperationRef: "/some/ref",
		OperationID:  "someOp",
		Description:  "some desc",
	}

	decl := Emit("testLink", link)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("links"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "var LinkTestLink") {
		t.Errorf("expected var LinkTestLink, got %s", out)
	}
}

func TestEmitLinkNil(t *testing.T) {
	if Emit("test", nil) != nil {
		t.Errorf("expected nil")
	}
}

func TestToPascalCase(t *testing.T) {
	if toPascalCase("") != "" {
		t.Errorf("expected empty string")
	}
	if toPascalCase("Test") != "Test" {
		t.Errorf("expected Test")
	}
	if toPascalCase("test") != "Test" {
		t.Errorf("expected Test")
	}
}
