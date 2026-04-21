package externaldocs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitExternalDocs(t *testing.T) {
	docs := &openapi.ExternalDocs{
		Description: "Find out more",
		URL:         "http://swagger.io",
	}

	decl, err := Emit(docs)
	if err != nil || decl == nil {
		t.Fatalf("unexpected error or nil decl: %v", err)
	}

	file := &dst.File{
		Name:  dst.NewIdent("externaldocs"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err = restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := strings.ReplaceAll(buf.String(), "\t", " ")
	out = strings.ReplaceAll(out, "\n", " ")
	for strings.Contains(out, "  ") {
		out = strings.ReplaceAll(out, "  ", " ")
	}

	if !strings.Contains(out, "var ExternalDocs = struct { Description string URL string }{") {
		t.Errorf("expected ExternalDocs struct")
	}
	if !strings.Contains(out, `Description: "Find out more"`) {
		t.Errorf("expected Description")
	}
	if !strings.Contains(out, `URL: "http://swagger.io"`) {
		t.Errorf("expected URL")
	}
}

func TestEmitExternalDocsNil(t *testing.T) {
	decl, err := Emit(nil)
	if err != nil || decl != nil {
		t.Errorf("expected nil")
	}
}
