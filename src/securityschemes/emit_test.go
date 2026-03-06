package securityschemes

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitSecurityScheme(t *testing.T) {
	scheme := &openapi.SecurityScheme{
		Type:             "http",
		Description:      "Basic Auth",
		Name:             "Authorization",
		In:               "header",
		Scheme:           "bearer",
		BearerFormat:     "JWT",
		OpenIDConnectURL: "http://example.com",
		Flows: &openapi.OAuthFlows{
			Implicit: &openapi.OAuthFlow{
				AuthorizationURL: "http://example.com/auth",
			},
		},
	}

	decl := Emit("bearerAuth", scheme)
	if decl == nil {
		t.Fatalf("expected decl")
	}

	file := &dst.File{
		Name:  dst.NewIdent("securityschemes"),
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

	if !strings.Contains(out, "var SecuritySchemeBearerAuth") {
		t.Errorf("expected var SecuritySchemeBearerAuth, got %s", out)
	}
	if !strings.Contains(out, `Type: "http"`) {
		t.Errorf("expected Type")
	}
	if !strings.Contains(out, `OpenIDConnectURL: "http://example.com"`) {
		t.Errorf("expected URL")
	}
	if !strings.Contains(out, `Flows:`) {
		t.Errorf("expected Flows")
	}
}

func TestEmitSecuritySchemeNil(t *testing.T) {
	if Emit("test", nil) != nil {
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
	if toPascalCase("test") != "Test" {
		t.Errorf("expected Test from test")
	}
}
