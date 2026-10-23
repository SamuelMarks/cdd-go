package mocks

import (
	"go/token"
	"testing"

	"github.com/dave/dst"
)

func TestParseExample(t *testing.T) {
	vs := &dst.ValueSpec{
		Names: []*dst.Ident{dst.NewIdent("MockUser")},
		Values: []dst.Expr{
			&dst.BasicLit{
				Kind:  token.STRING,
				Value: "`{\"id\": \"123\"}`",
			},
		},
		Decs: dst.ValueSpecDecorations{
			NodeDecs: dst.NodeDecs{
				Start: dst.Decorations{
					"// Test user",
					"// Detailed",
					"// @dataValue: {\"id\": \"123\"}",
					"// @serializedValue: id=123",
				},
			},
		},
	}

	ex, err := ParseExample(vs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ex.Summary != "Test user" {
		t.Errorf("expected Test user")
	}
	if ex.Description != "Detailed" {
		t.Errorf("expected Detailed")
	}
	if string(ex.DataValue) != `{"id": "123"}` {
		t.Errorf("expected DataValue, got %s", string(ex.DataValue))
	}
	if ex.SerializedValue != "id=123" {
		t.Errorf("expected SerializedValue, got %s", ex.SerializedValue)
	}

	if string(ex.Value) != `{"id": "123"}` {
		t.Errorf("expected json value")
	}
}

func TestParseExampleEmpty(t *testing.T) {
	vs := &dst.ValueSpec{
		Values: []dst.Expr{
			&dst.BasicLit{
				Kind:  token.STRING,
				Value: "`\"\"`",
			},
		},
	}
	ex, err := ParseExample(vs)
	if err != nil {
		t.Fatal(err)
	}
	if string(ex.Value) != "" {
		t.Errorf("expected empty string")
	}
}

func TestParseExampleNil(t *testing.T) {
	_, err := ParseExample(nil)
	if err == nil {
		t.Errorf("expected error")
	}
}
