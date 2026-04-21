package mocks

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/SamuelMarks/cdd-go/src/openapi"
	"github.com/dave/dst"
)

func TestEmitExample(t *testing.T) {
	ex := &openapi.Example{
		Summary:     "Test user",
		Description: "Detailed",
		Value:       json.RawMessage(`{"id": "123"}`),
	}

	decl, err := EmitExample("MockUser", ex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vs := decl.Specs[0].(*dst.ValueSpec)
	if vs.Names[0].Name != "MockUser" {
		t.Errorf("expected MockUser, got %s", vs.Names[0].Name)
	}

	if len(vs.Decs.Start) != 2 {
		t.Errorf("expected 2 line doc")
	} else if !strings.Contains(vs.Decs.Start[0], "Test user") {
		t.Errorf("expected Test user")
	}

	bl := vs.Values[0].(*dst.BasicLit)
	if bl.Value != "`{\"id\": \"123\"}`" {
		t.Errorf("expected json, got %s", bl.Value)
	}
}

func TestEmitExampleEmptyValue(t *testing.T) {
	ex := &openapi.Example{}
	decl, err := EmitExample("Empty", ex)
	if err != nil {
		t.Fatal(err)
	}
	vs := decl.Specs[0].(*dst.ValueSpec)
	bl := vs.Values[0].(*dst.BasicLit)
	if bl.Value != "`\"\"`" {
		t.Errorf("expected empty string, got %s", bl.Value)
	}
}

func TestEmitExampleNil(t *testing.T) {
	_, err := EmitExample("Mock", nil)
	if err == nil {
		t.Errorf("expected error")
	}
}
