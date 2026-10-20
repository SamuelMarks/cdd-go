// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestParameterValueUnion(t *testing.T) {
	sv := NewStringValue("hello")
	if sv.String() != "hello" {
		t.Errorf("expected hello, got %q", sv.String())
	}
	iv := NewIntValue(42)
	if iv.String() != "42" {
		t.Errorf("expected 42, got %q", iv.String())
	}
	fv := NewFloatValue(3.14)
	if fv.String() != "3.14" {
		t.Errorf("expected 3.14, got %q", fv.String())
	}
	bv := NewBoolValue(true)
	if bv.String() != "true" {
		t.Errorf("expected true, got %q", bv.String())
	}
	rv := ParameterValue{RawJSON: json.RawMessage(`{"k":"v"}`)}
	if rv.String() != `{"k":"v"}` {
		t.Errorf("expected raw json, got %q", rv.String())
	}
	empty := ParameterValue{}
	if empty.String() != "" {
		t.Errorf("expected empty string, got %q", empty.String())
	}

	for _, tc := range []ParameterValue{sv, iv, fv, bv, rv, empty} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		var out ParameterValue
		if err := json.Unmarshal(b, &out); err != nil {
			t.Fatalf("unmarshal failed for %s: %v", string(b), err)
		}
		if out.String() != tc.String() {
			t.Errorf("expected round-trip ParameterValue %q, got %q", tc.String(), out.String())
		}
	}

	var emptyUnmarshal ParameterValue
	if err := emptyUnmarshal.UnmarshalJSON([]byte("   ")); err != nil {
		t.Errorf("unexpected error on whitespace: %v", err)
	}
	if err := emptyUnmarshal.UnmarshalJSON([]byte(`"unterminated`)); err == nil {
		t.Error("expected error on invalid string JSON")
	}
	if err := emptyUnmarshal.UnmarshalJSON([]byte(`tru`)); err == nil {
		t.Error("expected error on invalid bool JSON")
	}
	if err := emptyUnmarshal.UnmarshalJSON([]byte(`{invalid`)); err == nil {
		t.Error("expected error on invalid object JSON")
	}
	if err := emptyUnmarshal.UnmarshalJSON([]byte(`999999999999999999999999999999999`)); err != nil {
		// Overflow int64 falls back to float64
		t.Errorf("expected fallback to float64, got err: %v", err)
	}
	if err := emptyUnmarshal.UnmarshalJSON([]byte(`invalid_number`)); err == nil {
		t.Error("expected error on invalid number")
	}
}

func TestDiscriminatedUnionsJSON(t *testing.T) {
	// ParameterOrReusable
	pInline := ParameterOrReusable{Parameter: &Parameter{Name: "id", In: ParamInPath, Value: NewIntValue(1)}}
	pReuse := ParameterOrReusable{Reusable: &ReusableObject{Reference: "$components.parameters.id"}}
	pNull := ParameterOrReusable{}

	for _, tc := range []ParameterOrReusable{pInline, pReuse, pNull} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}
		if len(b) > 0 && string(b) != "null" {
			var decoded ParameterOrReusable
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatal(err)
			}
		}
	}
	var badPR ParameterOrReusable
	if err := badPR.UnmarshalJSON([]byte(`{bad`)); err == nil {
		t.Error("expected error")
	}
	if err := badPR.UnmarshalJSON([]byte(`{"reference": "$components.parameters.x", "value": {"bad": [}}`)); err == nil {
		t.Error("expected error")
	}
	if err := badPR.UnmarshalJSON([]byte(`{"reference": "$components.parameters.x", "value": "valid"}`)); err != nil {
		t.Error(err)
	}
	if err := badPR.UnmarshalJSON([]byte(`{"name": 123}`)); err == nil {
		t.Error("expected error")
	}

	// SuccessActionOrReusable
	saInline := SuccessActionOrReusable{Action: &SuccessAction{Name: "done", Type: SuccessTypeEnd}}
	saReuse := SuccessActionOrReusable{Reusable: &ReusableObject{Reference: "$components.successActions.done"}}
	saNull := SuccessActionOrReusable{}
	for _, tc := range []SuccessActionOrReusable{saInline, saReuse, saNull} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != "null" {
			var decoded SuccessActionOrReusable
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatal(err)
			}
		}
	}
	var badSA SuccessActionOrReusable
	if err := badSA.UnmarshalJSON([]byte(`{bad`)); err == nil {
		t.Error("expected error")
	}
	if err := badSA.UnmarshalJSON([]byte(`{"reference": "$components.successActions.x", "value": "valid"}`)); err != nil {
		t.Error(err)
	}
	if err := badSA.UnmarshalJSON([]byte(`{"name": 123}`)); err == nil {
		t.Error("expected error")
	}

	// FailureActionOrReusable
	faInline := FailureActionOrReusable{Action: &FailureAction{Name: "err", Type: FailureTypeEnd}}
	faReuse := FailureActionOrReusable{Reusable: &ReusableObject{Reference: "$components.failureActions.err"}}
	faNull := FailureActionOrReusable{}
	for _, tc := range []FailureActionOrReusable{faInline, faReuse, faNull} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != "null" {
			var decoded FailureActionOrReusable
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatal(err)
			}
		}
	}
	var badFA FailureActionOrReusable
	if err := badFA.UnmarshalJSON([]byte(`{bad`)); err == nil {
		t.Error("expected error")
	}
	if err := badFA.UnmarshalJSON([]byte(`{"reference": "$components.failureActions.x", "value": "valid"}`)); err != nil {
		t.Error(err)
	}
	if err := badFA.UnmarshalJSON([]byte(`{"name": 123}`)); err == nil {
		t.Error("expected error")
	}

	// CriterionType
	var nilCT *CriterionType
	if nilCT.EffectiveType() != "simple" {
		t.Error("expected simple for nil CriterionType")
	}
	ctEmpty := &CriterionType{}
	if ctEmpty.EffectiveType() != "simple" {
		t.Error("expected simple for empty CriterionType")
	}
	ctSimple := CriterionType{SimpleType: "jsonpath"}
	if ctSimple.EffectiveType() != "jsonpath" {
		t.Error("expected jsonpath")
	}
	ctExpr := CriterionType{ExpressionType: &CriterionExpressionType{Type: "jsonpointer", Version: "rfc6901"}}
	if ctExpr.EffectiveType() != "jsonpointer" {
		t.Error("expected jsonpointer")
	}
	for _, tc := range []CriterionType{ctSimple, ctExpr} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}
		var decoded CriterionType
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, tc) {
			t.Errorf("expected %+v, got %+v", tc, decoded)
		}
	}
	var badCT CriterionType
	if err := badCT.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Error(err)
	}
	if err := badCT.UnmarshalJSON([]byte(`"unterminated`)); err == nil {
		t.Error("expected error")
	}
	if err := badCT.UnmarshalJSON([]byte(`{bad`)); err == nil {
		t.Error("expected error")
	}

	// OutputValue
	ovExpr := OutputValue{Expression: "$response.body#/id"}
	ovSel := OutputValue{Selector: &Selector{Context: "$response.body", Selector: "$.id", Type: &ctSimple}}
	for _, tc := range []OutputValue{ovExpr, ovSel} {
		b, err := json.Marshal(tc)
		if err != nil {
			t.Fatal(err)
		}
		var decoded OutputValue
		if err := json.Unmarshal(b, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, tc) {
			t.Errorf("expected %+v, got %+v", tc, decoded)
		}
	}
	var badOV OutputValue
	if err := badOV.UnmarshalJSON([]byte(`null`)); err != nil {
		t.Error(err)
	}
	if err := badOV.UnmarshalJSON([]byte(`"unterminated`)); err == nil {
		t.Error("expected error")
	}
	if err := badOV.UnmarshalJSON([]byte(`{bad`)); err == nil {
		t.Error("expected error")
	}
}

func TestParseAndValidate(t *testing.T) {
	if _, err := Parse(errReader{}); err == nil {
		t.Error("expected read error")
	}

	validJSON := `{
  "arazzo": "1.1.0",
  "$self": "https://example.com/arazzo.json",
  "x-root-ext": true,
  "info": {
    "title": "Petstore Workflows",
    "summary": "Login and buy pet",
    "description": "End-to-end pet adoption workflow",
    "version": "1.0.0",
    "x-info-ext": "meta"
  },
  "sourceDescriptions": [
    {
      "name": "petstore",
      "url": "./petstore.json",
      "type": "openapi",
      "x-sd-ext": 1
    }
  ],
  "components": {
    "parameters": {
      "sharedHeader": {
        "name": "X-Trace",
        "in": "header",
        "value": "$inputs.traceId"
      }
    },
    "successActions": {
      "endSuccess": {
        "name": "finish",
        "type": "end"
      }
    },
    "failureActions": {
      "retryOnce": {
        "name": "retry",
        "type": "retry",
        "retryAfter": 0.5,
        "retryLimit": 2
      }
    },
    "x-comp-ext": {}
  },
  "workflows": [
    {
      "workflowId": "authWorkflow",
      "summary": "Authenticate user",
      "steps": [
        {
          "stepId": "loginStep",
          "operationId": "loginUser",
          "parameters": [
            {
              "reference": "$components.parameters.sharedHeader"
            },
            {
              "name": "username",
              "in": "query",
              "value": "$inputs.username"
            }
          ],
          "successCriteria": [
            {
              "condition": "$statusCode == 200",
              "type": "simple"
            }
          ],
          "onSuccess": [
            {
              "reference": "$components.successActions.endSuccess"
            }
          ],
          "onFailure": [
            {
              "reference": "$components.failureActions.retryOnce"
            }
          ],
          "outputs": {
            "token": "$response.body#/token",
            "userId": {
              "context": "$response.body",
              "selector": "$.user.id",
              "type": {
                "type": "jsonpath",
                "version": "rfc9535"
              }
            }
          },
          "x-step-ext": "step"
        }
      ],
      "x-wf-ext": "wf"
    },
    {
      "workflowId": "buyPet",
      "dependsOn": ["authWorkflow"],
      "parameters": [
        {
          "reference": "$components.parameters.sharedHeader"
        }
      ],
      "successActions": [
        {
          "reference": "$components.successActions.endSuccess"
        }
      ],
      "failureActions": [
        {
          "reference": "$components.failureActions.retryOnce"
        }
      ],
      "steps": [
        {
          "stepId": "createPet",
          "operationPath": "{$sourceDescriptions.petstore.url}#/paths/~1pet/post",
          "requestBody": {
            "contentType": "application/json",
            "payload": {"name": "doggie", "status": "available"},
            "replacements": [
              {
                "target": "/name",
                "value": "$inputs.petName"
              }
            ]
          }
        },
        {
          "stepId": "notifyAsync",
          "channelPath": "{$sourceDescriptions.petstore.url}#/channels/petCreated",
          "action": "send",
          "correlationId": "$message.header.correlationId",
          "timeout": 5000,
          "dependsOn": ["createPet"]
        },
        {
          "stepId": "subWorkflowStep",
          "workflowId": "authWorkflow",
          "dependsOn": ["notifyAsync"]
        }
      ]
    }
  ]
}`

	doc, err := Parse(strings.NewReader(validJSON))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if err := Validate(doc); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if len(doc.Extensions) == 0 || len(doc.Info.Extensions) == 0 {
		t.Error("expected root and info extensions to be captured")
	}

	// Test YAML parsing round-trip
	var yamlBuf bytes.Buffer
	if err := EmitYAML(&yamlBuf, doc); err != nil {
		t.Fatalf("EmitYAML failed: %v", err)
	}
	docFromYAML, err := ParseBytes(yamlBuf.Bytes())
	if err != nil {
		t.Fatalf("ParseBytes from YAML failed: %v", err)
	}
	if err := Validate(docFromYAML); err != nil {
		t.Fatalf("Validate from YAML failed: %v", err)
	}

	// Test JSON emission
	var jsonBuf bytes.Buffer
	if err := Emit(&jsonBuf, doc); err != nil {
		t.Fatalf("Emit JSON failed: %v", err)
	}

	// Test invalid inputs to ParseBytes
	if _, err := ParseBytes([]byte(`{}` /* missing arazzo version */)); err == nil {
		t.Error("expected error for missing arazzo field")
	}
	if _, err := ParseBytes([]byte(`::invalid yaml and json:: [`)); err == nil {
		t.Error("expected error for invalid syntax")
	}
	populateDocumentExtensions([]byte(`{invalid`), doc)
}

func TestValidateErrors(t *testing.T) {
	if err := Validate(nil); err == nil {
		t.Error("expected error on nil doc")
	}

	baseDoc := func() *Arazzo {
		return &Arazzo{
			Arazzo: "1.1.0",
			Info:   Info{Title: "T", Version: "1.0.0"},
			SourceDescriptions: []SourceDescription{
				{Name: "api", URL: "./openapi.json", Type: SourceTypeOpenAPI},
			},
			Workflows: []Workflow{
				{
					WorkflowID: "wf1",
					Steps: []Step{
						{StepID: "s1", OperationID: "op1"},
					},
				},
			},
		}
	}

	t.Run("bad version", func(t *testing.T) {
		d := baseDoc()
		d.Arazzo = "2.0.0"
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty title", func(t *testing.T) {
		d := baseDoc()
		d.Info.Title = ""
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty version", func(t *testing.T) {
		d := baseDoc()
		d.Info.Version = ""
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("no sourceDescriptions", func(t *testing.T) {
		d := baseDoc()
		d.SourceDescriptions = nil
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty sourceDescription name", func(t *testing.T) {
		d := baseDoc()
		d.SourceDescriptions[0].Name = ""
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("invalid sourceDescription type", func(t *testing.T) {
		d := baseDoc()
		d.SourceDescriptions[0].Type = "graphql"
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("duplicate sourceDescription name", func(t *testing.T) {
		d := baseDoc()
		d.SourceDescriptions = append(d.SourceDescriptions, d.SourceDescriptions[0])
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("no workflows", func(t *testing.T) {
		d := baseDoc()
		d.Workflows = nil
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("empty workflowId", func(t *testing.T) {
		d := baseDoc()
		d.Workflows[0].WorkflowID = ""
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("duplicate workflowId", func(t *testing.T) {
		d := baseDoc()
		d.Workflows = append(d.Workflows, d.Workflows[0])
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("unknown workflow dependsOn", func(t *testing.T) {
		d := baseDoc()
		d.Workflows[0].DependsOn = []string{"nonexistent"}
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("workflow cycle", func(t *testing.T) {
		d := baseDoc()
		d.Workflows[0].DependsOn = []string{"wf2"}
		d.Workflows = append(d.Workflows, Workflow{
			WorkflowID: "wf2",
			DependsOn:  []string{"wf1"},
			Steps:      []Step{{StepID: "s1", OperationID: "op1"}},
		})
		if Validate(d) == nil {
			t.Error("expected cycle error")
		}
	})
	t.Run("workflow parameter & action errors", func(t *testing.T) {
		d := baseDoc()
		d.Workflows[0].Parameters = []ParameterOrReusable{{Parameter: &Parameter{Name: ""}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].SuccessActions = []SuccessActionOrReusable{{Action: &SuccessAction{Name: ""}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].FailureActions = []FailureActionOrReusable{{Action: &FailureAction{Name: ""}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}
	})
	t.Run("step validation errors", func(t *testing.T) {
		d := baseDoc()
		d.Workflows[0].Steps = nil
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].StepID = ""
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps = append(d.Workflows[0].Steps, d.Workflows[0].Steps[0])
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OperationID = ""
		if Validate(d) == nil {
			t.Error("expected error on 0 targets")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].Parameters = []ParameterOrReusable{{Reusable: &ReusableObject{Reference: "badref"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].Parameters = []ParameterOrReusable{{Reusable: &ReusableObject{Reference: "$components.parameters.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Components = &Components{Parameters: map[string]Parameter{}}
		d.Workflows[0].Steps[0].Parameters = []ParameterOrReusable{{Reusable: &ReusableObject{Reference: "$components.parameters.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnSuccess = []SuccessActionOrReusable{{Reusable: &ReusableObject{Reference: "badref"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnSuccess = []SuccessActionOrReusable{{Reusable: &ReusableObject{Reference: "$components.successActions.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Components = &Components{SuccessActions: map[string]SuccessAction{}}
		d.Workflows[0].Steps[0].OnSuccess = []SuccessActionOrReusable{{Reusable: &ReusableObject{Reference: "$components.successActions.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnSuccess = []SuccessActionOrReusable{{Action: &SuccessAction{Name: "x", Type: "invalid"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnSuccess = []SuccessActionOrReusable{{Action: &SuccessAction{Name: "ok", Type: SuccessTypeEnd}}}
		if err := Validate(d); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnFailure = []FailureActionOrReusable{{Reusable: &ReusableObject{Reference: "badref"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnFailure = []FailureActionOrReusable{{Reusable: &ReusableObject{Reference: "$components.failureActions.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Components = &Components{FailureActions: map[string]FailureAction{}}
		d.Workflows[0].Steps[0].OnFailure = []FailureActionOrReusable{{Reusable: &ReusableObject{Reference: "$components.failureActions.missing"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnFailure = []FailureActionOrReusable{{Action: &FailureAction{Name: "x", Type: "invalid"}}}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].OnFailure = []FailureActionOrReusable{{Action: &FailureAction{Name: "err", Type: FailureTypeEnd}}}
		if err := Validate(d); err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		d = baseDoc()
		d.Workflows[0].Steps[0].DependsOn = []string{"unknownStep"}
		if Validate(d) == nil {
			t.Error("expected error")
		}

		d = baseDoc()
		d.Workflows[0].Steps = []Step{
			{StepID: "s1", OperationID: "op1", DependsOn: []string{"s2"}},
			{StepID: "s2", OperationID: "op2", DependsOn: []string{"s1"}},
		}
		if Validate(d) == nil {
			t.Error("expected step cycle error")
		}
	})
}
