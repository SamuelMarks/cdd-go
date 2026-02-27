package openapi

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseAndEmit(t *testing.T) {
	input := `{
  "openapi": "3.2.0",
  "info": {
    "title": "Example API",
    "version": "1.0.1"
  }
}
`

	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if oa.OpenAPI != "3.2.0" {
		t.Errorf("expected openapi version 3.2.0, got %s", oa.OpenAPI)
	}

	if oa.Info.Title != "Example API" {
		t.Errorf("expected title 'Example API', got %s", oa.Info.Title)
	}

	if oa.Info.Version != "1.0.1" {
		t.Errorf("expected version '1.0.1', got %s", oa.Info.Version)
	}

	var buf bytes.Buffer
	err = Emit(&buf, oa)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if buf.String() != input {
		t.Errorf("expected emitted string to match input.\nExpected:\n%s\nGot:\n%s", input, buf.String())
	}
}

func TestParseError(t *testing.T) {
	input := "{invalid_json}"
	r := strings.NewReader(input)
	_, err := Parse(r)
	if err == nil {
		t.Error("expected error parsing invalid json, got nil")
	}
}
