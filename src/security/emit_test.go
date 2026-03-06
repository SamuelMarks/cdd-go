package security

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samuel/cdd-go/src/openapi"
)

func TestEmitSecurity(t *testing.T) {
	security := []openapi.SecurityRequirement{
		{
			"api_key": {},
		},
		{
			"oauth2": {"read:users", "write:users"},
		},
	}

	cl := Emit(security)
	if cl == nil {
		t.Fatalf("expected composite lit")
	}

	decl := &dst.GenDecl{
		Tok: 114, // token.VAR
		Specs: []dst.Spec{
			&dst.ValueSpec{
				Names:  []*dst.Ident{dst.NewIdent("Security")},
				Values: []dst.Expr{cl},
			},
		},
	}

	file := &dst.File{
		Name:  dst.NewIdent("security"),
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

	if !strings.Contains(out, `"api_key": []string{}`) {
		t.Errorf("expected api_key, got %s", out)
	}
	if !strings.Contains(out, `"oauth2": []string{"read:users", "write:users"}`) {
		t.Errorf("expected oauth2 scopes, got %s", out)
	}
}

func TestEmitSecurityEmpty(t *testing.T) {
	if Emit(nil) != nil {
		t.Errorf("expected nil")
	}
	if Emit([]openapi.SecurityRequirement{}) != nil {
		t.Errorf("expected nil")
	}
}
