package openapi

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failingReader struct{}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}

func TestParseYAML(t *testing.T) {
	input := `
openapi: 3.2.0
info:
  title: Example API YAML
  version: 1.0.1
`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error for yaml, got %v", err)
	}
	if oa.OpenAPI != "3.2.0" {
		t.Errorf("expected 3.2.0, got %v", oa.OpenAPI)
	}
}

func TestParseReadAllError(t *testing.T) {
	_, err := Parse(failingReader{})
	if err == nil {
		t.Fatalf("expected error from Parse, got nil")
	}
}

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
	input := "{invalid_json: "
	r := strings.NewReader(input)
	_, err := Parse(r)
	if err == nil {
		t.Error("expected error parsing invalid json, got nil")
	}
}

func TestFullSchemaCoverage(t *testing.T) {
	input := `{
  "openapi": "3.2.0",
  "info": {
    "title": "Full Schema API",
    "version": "1.0.0"
  },
  "components": {
    "schemas": {
      "Pet": {
        "type": "object",
        "required": [
          "id",
          "name"
        ],
        "discriminator": {
          "propertyName": "petType",
          "mapping": {
            "dog": "#/components/schemas/Dog"
          }
        },
        "xml": {
          "name": "pet",
          "wrapped": true
        },
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64"
          },
          "name": {
            "type": "string"
          },
          "petType": {
            "type": "string"
          }
        }
      }
    }
  }
}
`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	petSchema := oa.Components.Schemas["Pet"]
	if petSchema.Discriminator == nil || petSchema.Discriminator.PropertyName != "petType" {
		t.Errorf("expected discriminator with propertyName petType")
	}

	if petSchema.XML == nil || petSchema.XML.Name != "pet" || petSchema.XML.Wrapped != true {
		t.Errorf("expected xml with name pet and wrapped true")
	}

	var buf bytes.Buffer
	err = Emit(&buf, oa)
	if err != nil {
		t.Fatalf("expected no error emitting full schema")
	}

	if !strings.Contains(buf.String(), "discriminator") {
		t.Errorf("expected output to contain discriminator")
	}
}

func TestParseJSONDefinitions(t *testing.T) {
	input := `{
  "swagger": "2.0",
  "definitions": {
    "Pet": {
      "type": "object"
    }
  }
}`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oa.Components == nil || len(oa.Components.Schemas) == 0 {
		t.Errorf("expected components.schemas to be populated from definitions")
	}
}

func TestParseYAMLDefinitions(t *testing.T) {
	input := `
swagger: "2.0"
definitions:
  Pet:
    type: object
`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oa.Components == nil || len(oa.Components.Schemas) == 0 {
		t.Errorf("expected components.schemas to be populated from definitions")
	}
}

func TestParseYAMLDefinitionsWithExistingComponents(t *testing.T) {
	input := `
swagger: "2.0"
components:
  responses:
    Success:
      description: "OK"
definitions:
  Pet:
    type: object
`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oa.Components == nil || len(oa.Components.Schemas) == 0 {
		t.Errorf("expected components.schemas to be populated from definitions")
	}
	if len(oa.Components.Responses) == 0 {
		t.Errorf("expected components.responses to be preserved")
	}
}

func TestParseJSONDefinitionsWithExistingComponents(t *testing.T) {
	input := `{
  "swagger": "2.0",
  "components": {
    "responses": {
      "Success": {
        "description": "OK"
      }
    }
  },
  "definitions": {
    "Pet": {
      "type": "object"
    }
  }
}`
	r := strings.NewReader(input)
	oa, err := Parse(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if oa.Components == nil || len(oa.Components.Schemas) == 0 {
		t.Errorf("expected components.schemas to be populated from definitions")
	}
	if len(oa.Components.Responses) == 0 {
		t.Errorf("expected components.responses to be preserved")
	}
}

func TestPathItemOperationsAndSetOperation(t *testing.T) {
	if ops := PathItemOperations(nil); ops != nil {
		t.Errorf("expected nil for nil PathItem, got %v", ops)
	}
	var pi PathItem
	verbs := []string{"get", "post", "put", "delete", "patch", "options", "head", "trace"}
	for _, v := range verbs {
		pi.SetOperation(v, &Operation{OperationID: v + "Op"})
	}
	ops := PathItemOperations(&pi)
	if len(ops) != 8 {
		t.Fatalf("expected 8 operations, got %d", len(ops))
	}
	for i, v := range verbs {
		if ops[i].Method != v || ops[i].Operation.OperationID != v+"Op" {
			t.Errorf("expected method %s with op %sOp, got %+v", v, v, ops[i])
		}
	}
}

func TestSchemaUnmarshalJSONUnions(t *testing.T) {
	input := `{
  "openapi": "3.2.0",
  "components": {
    "schemas": {
      "NullableStr": {
        "type": ["null", "string"],
        "additionalProperties": false
      },
      "OnlyNull": {
        "type": ["null"],
        "additionalProperties": {"type": "integer"}
      }
    }
  }
}`
	oa, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error parsing OpenAPI 3.2.0 schema unions: %v", err)
	}
	ns := oa.Components.Schemas["NullableStr"]
	if ns.Type != "string" || len(ns.TypeArray) != 2 || ns.AdditionalPropertiesBool == nil || *ns.AdditionalPropertiesBool != false {
		t.Errorf("unexpected NullableStr schema: %+v", ns)
	}
	on := oa.Components.Schemas["OnlyNull"]
	if on.Type != "null" || on.AdditionalProperties == nil || on.AdditionalProperties.Type != "integer" {
		t.Errorf("unexpected OnlyNull schema: %+v", on)
	}

	var badType Schema
	if err := badType.UnmarshalJSON([]byte(`{"type": 123}`)); err == nil {
		t.Errorf("expected error for invalid type number")
	}
	var badAddProps Schema
	if err := badAddProps.UnmarshalJSON([]byte(`{"additionalProperties": 123}`)); err == nil {
		t.Errorf("expected error for invalid additionalProperties number")
	}
	var badJSON Schema
	if err := badJSON.UnmarshalJSON([]byte(`{invalid`)); err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}
