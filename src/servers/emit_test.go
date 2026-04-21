package servers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitServers(t *testing.T) {
	servers := []openapi.Server{
		{URL: "https://api.example.com", Description: "Prod"},
		{URL: "http://localhost:8080"},
	}

	decl := Emit(servers)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("servers"),
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

	if !strings.Contains(out, "var Servers = []struct { URL string Description string }{") {
		t.Errorf("expected Servers struct array")
	}
	if !strings.Contains(out, `{URL: "https://api.example.com", Description: "Prod"}`) {
		t.Errorf("expected prod server")
	}
	if !strings.Contains(out, `{URL: "http://localhost:8080"}`) {
		t.Errorf("expected localhost server")
	}
}

func TestEmitServersEmpty(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
	if Emit([]openapi.Server{}) != nil {
		t.Errorf("expected nil for empty slice")
	}
}
