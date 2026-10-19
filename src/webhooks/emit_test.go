package webhooks

import (
	"bytes"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitWebhook(t *testing.T) {
	wh := map[string]openapi.PathItem{
		"newPet": {Summary: "Test"},
	}

	decl := Emit("testWebhook", wh)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("webhooks"),
		Decls: []dst.Decl{decl},
	}

	restorer := decorator.NewRestorer()
	var buf bytes.Buffer
	err := restorer.Fprint(&buf, file)
	if err != nil {
		t.Fatalf("unexpected print error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "var WebhookTestWebhook") {
		t.Errorf("expected var WebhookTestWebhook, got %s", out)
	}
}

func TestEmitWebhookNil(t *testing.T) {
	if Emit("test", nil) != nil {
		t.Errorf("expected nil")
	}
}
