// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write error")
}

func TestEmitAndMarshalDocument(t *testing.T) {
	if _, err := MarshalDocument(nil); err == nil {
		t.Error("expected error on nil doc")
	}

	if err := Emit(errWriter{}, nil); err == nil {
		t.Error("expected error on nil doc")
	}

	if err := EmitYAML(errWriter{}, nil); err == nil {
		t.Error("expected error on nil doc")
	}

	doc := &Arazzo{
		Arazzo: "1.1.0",
		Info: Info{
			Title:      "Test",
			Version:    "1.0.0",
			Extensions: map[string]json.RawMessage{"x-info": json.RawMessage(`"val"`)},
		},
		SourceDescriptions: []SourceDescription{
			{Name: "api", URL: "./openapi.json", Type: SourceTypeOpenAPI},
		},
		Workflows: []Workflow{
			{
				WorkflowID: "wf",
				Steps:      []Step{{StepID: "s1", OperationID: "op1"}},
			},
		},
		Extensions: map[string]json.RawMessage{"x-root": json.RawMessage(`true`)},
	}

	if err := Emit(errWriter{}, doc); err == nil {
		t.Error("expected write error")
	}
	if err := EmitYAML(errWriter{}, doc); err == nil {
		t.Error("expected write error")
	}

	var buf bytes.Buffer
	if err := Emit(&buf, doc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test invalid rawjson in ParameterValue causing json.Marshal error
	badDoc := &Arazzo{
		Arazzo: "1.1.0",
		Info:   Info{Title: "T", Version: "1.0.0"},
		Workflows: []Workflow{
			{
				WorkflowID: "wf",
				Steps: []Step{
					{
						StepID:      "s1",
						OperationID: "op1",
						Parameters: []ParameterOrReusable{
							{Parameter: &Parameter{Name: "p", Value: ParameterValue{RawJSON: json.RawMessage(`{invalid`)}}},
						},
					},
				},
			},
		},
	}
	if _, err := MarshalDocument(badDoc); err == nil {
		t.Error("expected marshal error on invalid RawJSON")
	}

	badExtDoc := &Arazzo{
		Arazzo:     "1.1.0",
		Info:       Info{Title: "T", Version: "1.0.0"},
		Extensions: map[string]json.RawMessage{"x-bad": json.RawMessage(`{invalid`)},
	}
	if _, err := MarshalDocument(badExtDoc); err == nil {
		t.Error("expected marshal error on invalid extension RawJSON")
	}
}
