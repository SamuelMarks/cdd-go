package tests

import (
	"testing"

	"github.com/dave/dst"
)

func TestParseTest(t *testing.T) {
	fd := &dst.FuncDecl{
		Name: dst.NewIdent("TestGetUser"),
		Decs: dst.FuncDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// TestGetUser tests the Get User operation.",
					"// Other desc",
				},
			},
		},
	}

	op, err := ParseTest(fd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if op.OperationID != "getUser" {
		t.Errorf("expected getUser, got %s", op.OperationID)
	}

	if op.Summary != "Get User" {
		t.Errorf("expected Get User, got %s", op.Summary)
	}
	if op.Description != "Other desc" {
		t.Errorf("expected Other desc, got %s", op.Description)
	}
}

func TestParseTestNotTest(t *testing.T) {
	fd := &dst.FuncDecl{
		Name: dst.NewIdent("GetUser"),
	}
	_, err := ParseTest(fd)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseTestNil(t *testing.T) {
	_, err := ParseTest(nil)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseTestNoSummaryMatch(t *testing.T) {
	fd := &dst.FuncDecl{
		Name: dst.NewIdent("TestSomething"),
		Decs: dst.FuncDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// Something does something",
				},
			},
		},
	}

	op, err := ParseTest(fd)
	if err != nil {
		t.Fatal(err)
	}
	if op.Summary != "Something does something" {
		t.Errorf("expected summary, got %s", op.Summary)
	}
}
