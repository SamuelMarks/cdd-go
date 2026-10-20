package functions

import (
	"testing"

	"github.com/dave/dst"
)

func TestParseOperation(t *testing.T) {
	fd := &dst.FuncDecl{
		Name: dst.NewIdent("CreateUser"),
		Type: &dst.FuncType{
			Params: &dst.FieldList{
				List: []*dst.Field{
					{Names: []*dst.Ident{dst.NewIdent("id")}, Type: dst.NewIdent("string")},
					{Names: []*dst.Ident{dst.NewIdent("body")}, Type: dst.NewIdent("string")},
				},
			},
			Results: &dst.FieldList{
				List: []*dst.Field{
					{Type: dst.NewIdent("string")},
					{Type: dst.NewIdent("error")},
				},
			},
		},
		Decs: dst.FuncDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// Path: /users",
					"// METHOD: POST",
					"// Creates a new user",
					"// Detailed description",
					"// is here",
				},
			},
		},
	}

	op, err := ParseOperation(fd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if op.OperationID != "CreateUser" {
		t.Errorf("expected CreateUser, got %s", op.OperationID)
	}

	if op.Summary != "Creates a new user" {
		t.Errorf("expected summary, got %s", op.Summary)
	}

	expectedDesc := "Detailed description\nis here"
	if op.Description != expectedDesc {
		t.Errorf("expected description %q, got %q", expectedDesc, op.Description)
	}
}

func TestParseOperationNoDoc(t *testing.T) {
	fd := &dst.FuncDecl{
		Name: dst.NewIdent("SimpleFunc"),
	}

	op, err := ParseOperation(fd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.OperationID != "SimpleFunc" {
		t.Errorf("expected SimpleFunc")
	}
	if op.Summary != "" || op.Description != "" {
		t.Errorf("expected no doc")
	}
}

func TestParseOperationNil(t *testing.T) {
	_, err := ParseOperation(nil)
	if err == nil {
		t.Errorf("expected error")
	}
}
