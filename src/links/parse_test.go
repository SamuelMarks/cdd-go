package links

import (
	"go/token"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestParseLink(t *testing.T) {
	link := &openapi.Link{
		OperationRef: "/test",
		OperationID:  "testOp",
		Description:  "Test",
	}

	decl := Emit("testLink", link)

	name, parsedLink := Parse(decl)
	if name != "testLink" {
		t.Errorf("expected name testLink, got %s", name)
	}
	if parsedLink.OperationID != "testOp" {
		t.Errorf("expected operationID testOp, got %s", parsedLink.OperationID)
	}
}

func TestParseLinkNil(t *testing.T) {
	name, link := Parse(nil)
	if name != "" || link != nil {
		t.Errorf("expected nil, got %s, %v", name, link)
	}
}

func TestParseLinkEmptyDecl(t *testing.T) {
	decl := &dst.GenDecl{Tok: token.VAR}
	name, link := Parse(decl)
	if name != "" || link != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.TypeSpec{}}
	name, link = Parse(decl)
	if name != "" || link != nil {
		t.Errorf("expected nil")
	}

	decl.Specs = []dst.Spec{&dst.ValueSpec{Names: []*dst.Ident{dst.NewIdent("LinkTest")}, Values: []dst.Expr{&dst.BasicLit{}}}}
	name, link = Parse(decl)
	if name != "" || link != nil {
		t.Errorf("expected nil")
	}
}
