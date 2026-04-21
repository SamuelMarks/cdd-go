package tags

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitTags(t *testing.T) {
	tags := []openapi.Tag{
		{Name: "users", Description: "Operations about users"},
		{Name: "posts"},
	}

	decl := Emit(tags)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("tags"),
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

	if !strings.Contains(out, "var Tags = []struct { Name string Description string }{") {
		t.Errorf("expected Tags struct array")
	}
	if !strings.Contains(out, `{Name: "users", Description: "Operations about users"}`) {
		t.Errorf("expected users tag")
	}
	if !strings.Contains(out, `{Name: "posts"}`) {
		t.Errorf("expected posts tag")
	}
}

func TestEmitTagsEmpty(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
	if Emit([]openapi.Tag{}) != nil {
		t.Errorf("expected nil for empty slice")
	}
}
