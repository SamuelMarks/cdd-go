package webhooks

import (
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestParseWebhook(t *testing.T) {
	wh := map[string]openapi.PathItem{
		"newPet": {Summary: "Test"},
	}

	decl := Emit("testWebhook", wh)

	name, parsedWebhook := Parse(decl)
	if name != "testWebhook" {
		t.Errorf("expected name testWebhook, got %s", name)
	}
	if parsedWebhook["newPet"].Summary != "Test" {
		t.Errorf("expected summary Test")
	}
}

func TestParseWebhookNil(t *testing.T) {
	name, wh := Parse(nil)
	if name != "" || wh != nil {
		t.Errorf("expected nil, got %s, %v", name, wh)
	}
}

func TestParseWebhookEmptyDecl(t *testing.T) {
	decl := &dst.GenDecl{Tok: token.VAR}
	name, wh := Parse(decl)
	if name != "" || wh != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.TypeSpec{}}
	name, wh = Parse(decl)
	if name != "" || wh != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("WebhookTest")}, Values: []dst.Expr{&dst.BasicLit{}}}}
	name, wh = Parse(decl)
	if name != "" || wh != nil {
		t.Errorf("expected nil")
	}
}
