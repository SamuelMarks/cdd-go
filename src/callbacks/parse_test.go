package callbacks

import (
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestParseCallback(t *testing.T) {
	cb := openapi.Callback{
		"http://test.com": openapi.PathItem{Summary: "Test"},
	}

	decl := Emit("testCallback", &cb)

	name, parsedCallback := Parse(decl)
	if name != "testCallback" {
		t.Errorf("expected name testCallback, got %s", name)
	}
	if (*parsedCallback)["http://test.com"].Summary != "Test" {
		t.Errorf("expected summary Test")
	}
}

func TestParseCallbackNil(t *testing.T) {
	name, cb := Parse(nil)
	if name != "" || cb != nil {
		t.Errorf("expected nil, got %s, %v", name, cb)
	}
}

func TestParseCallbackEmptyDecl(t *testing.T) {
	decl := &dst.GenDecl{Tok: token.VAR}
	name, cb := Parse(decl)
	if name != "" || cb != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.TypeSpec{}}
	name, cb = Parse(decl)
	if name != "" || cb != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("CallbackTest")}, Values: []dst.Expr{&dst.BasicLit{}}}}
	name, cb = Parse(decl)
	if name != "" || cb != nil {
		t.Errorf("expected nil")
	}
}
