package oauthflows

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

func TestEmitOAuthFlows(t *testing.T) {
	flows := &openapi.OAuthFlows{
		Implicit: &openapi.OAuthFlow{
			AuthorizationURL: "https://example.com/api/oauth/dialog",
			Scopes: map[string]string{
				"write:pets": "modify pets in your account",
				"read:pets":  "read your pets",
			},
		},
		AuthorizationCode: &openapi.OAuthFlow{
			AuthorizationURL: "https://example.com/api/oauth/dialog",
			TokenURL:         "https://example.com/api/oauth/token",
			RefreshURL:       "https://example.com/api/oauth/refresh",
		},
		Password: &openapi.OAuthFlow{
			TokenURL: "https://example.com/api/oauth/token",
		},
		ClientCredentials: &openapi.OAuthFlow{
			TokenURL: "https://example.com/api/oauth/token",
		},
	}

	cl := Emit(flows)
	if cl == nil {
		t.Fatalf("expected emit")
	}

	decl := &dst.GenDecl{
		Tok: token.VAR,
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Flows")},
				Values: []dst.Expr{cl},
			},
		},
	}

	file := &dst.File{
		Name:  dst.NewIdent("oauthflows"),
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

	if !strings.Contains(out, `Implicit: OAuthFlow{`) {
		t.Errorf("expected Implicit flow")
	}
	if !strings.Contains(out, `AuthorizationURL: "https://example.com/api/oauth/dialog"`) {
		t.Errorf("expected Auth URL")
	}
	if !strings.Contains(out, `Scopes: map[string]string{`) {
		t.Errorf("expected Scopes map")
	}
	if !strings.Contains(out, `AuthorizationCode: OAuthFlow{`) {
		t.Errorf("expected AuthorizationCode flow")
	}
	if !strings.Contains(out, `Password: OAuthFlow{`) {
		t.Errorf("expected Password flow")
	}
	if !strings.Contains(out, `ClientCredentials: OAuthFlow{`) {
		t.Errorf("expected ClientCredentials flow")
	}
}

func TestEmitOAuthFlowsNil(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
	if Emit(&openapi.OAuthFlows{}) != nil {
		t.Errorf("expected nil")
	}
}
