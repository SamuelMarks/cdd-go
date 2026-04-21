package info

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitInfo(t *testing.T) {
	info := openapi.Info{
		Title:          "Test API",
		Version:        "1.0.0",
		Description:    "A test API",
		Summary:        "Test Summary",
		TermsOfService: "http://example.com/terms",
		Contact: &openapi.Contact{
			Name:  "Test Contact",
			URL:   "http://example.com/contact",
			Email: "test@example.com",
		},
		License: &openapi.License{
			Name:       "MIT",
			URL:        "http://example.com/license",
			Identifier: "MIT",
		},
	}

	decl := Emit(info)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("info"),
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

	if !strings.Contains(out, "const Info = struct") {
		t.Errorf("expected const Info")
	}
	if !strings.Contains(out, `Title: "Test API"`) {
		t.Errorf("expected Title")
	}
	if !strings.Contains(out, `Contact: {Name: "Test Contact", URL: "http://example.com/contact", Email: "test@example.com"}`) {
		t.Errorf("expected Contact")
	}
	if !strings.Contains(out, `License: {Name: "MIT", URL: "http://example.com/license", Identifier: "MIT"}`) {
		t.Errorf("expected License")
	}
}

func TestEmitInfoNil(t *testing.T) {
	if Emit(openapi.Info{}) != nil {
		t.Errorf("expected nil")
	}
}

func TestEmitInfoEmpty(t *testing.T) {
	info := openapi.Info{Contact: &openapi.Contact{}, License: &openapi.License{}}
	if Emit(info) != nil {
		t.Errorf("expected nil for empty nested structs")
	}
}
