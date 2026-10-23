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

	exData := &openapi.Example{
		DataValue:       json.RawMessage(`{"x":1}`),
		SerializedValue: "x=1",
	}
	declData, err := EmitExample("DataEx", exData)
	if err != nil {
		t.Fatal(err)
	}
	vsData := declData.Specs[0].(*dst.ValueSpec)
	if vsData.Values[0].(*dst.BasicLit).Value != "`{\"x\":1}`" {
		t.Errorf("expected DataValue literal, got %s", vsData.Values[0].(*dst.BasicLit).Value)
	}

	exSer := &openapi.Example{
		SerializedValue: "x=1",
	}
	declSer, err := EmitExample("SerEx", exSer)
	if err != nil {
		t.Fatal(err)
	}
	vsSer := declSer.Specs[0].(*dst.ValueSpec)
	if vsSer.Values[0].(*dst.BasicLit).Value != "`x=1`" {
		t.Errorf("expected SerializedValue literal, got %s", vsSer.Values[0].(*dst.BasicLit).Value)
	}
}

func TestEmitExampleNil(t *testing.T) {
	_, err := EmitExample("Mock", nil)
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestEmitExampleInvalidIdentifier(t *testing.T) {
	_, err := EmitExample("123-invalid", &openapi.Example{})
	if err == nil {
		t.Errorf("expected error for invalid Go identifier")
	}
	_, err = EmitExample("", &openapi.Example{})
	if err == nil {
		t.Errorf("expected error for empty identifier")
	}
}
