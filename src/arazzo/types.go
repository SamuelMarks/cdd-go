// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// SourceDescriptionType represents the specification protocol type of a SourceDescription.
type SourceDescriptionType string

const (
	// SourceTypeOpenAPI indicates an OpenAPI Specification source description.
	SourceTypeOpenAPI SourceDescriptionType = "openapi"
	// SourceTypeAsyncAPI indicates an AsyncAPI Specification source description.
	SourceTypeAsyncAPI SourceDescriptionType = "asyncapi"
	// SourceTypeArazzo indicates an external Arazzo Specification source description.
	SourceTypeArazzo SourceDescriptionType = "arazzo"
)

// StepActionType represents the AsyncAPI message flow intent for a Step in Arazzo 1.1.0.
type StepActionType string

const (
	// StepActionSend indicates publishing or sending a message to a channel.
	StepActionSend StepActionType = "send"
	// StepActionReceive indicates subscribing to or receiving a message from a channel.
	StepActionReceive StepActionType = "receive"
)

// ParameterIn represents the target location of an Arazzo Parameter.
type ParameterIn string

const (
	// ParamInPath indicates a URL path parameter.
	ParamInPath ParameterIn = "path"
	// ParamInQuery indicates a URL query parameter.
	ParamInQuery ParameterIn = "query"
	// ParamInQueryString indicates a raw URL querystring parameter (Arazzo 1.1.0).
	ParamInQueryString ParameterIn = "querystring"
	// ParamInHeader indicates an HTTP header parameter.
	ParamInHeader ParameterIn = "header"
	// ParamInCookie indicates an HTTP cookie parameter.
	ParamInCookie ParameterIn = "cookie"
)

// SuccessActionType represents the control flow primitive of a SuccessAction.
type SuccessActionType string

const (
	// SuccessTypeEnd terminates the workflow upon step success.
	SuccessTypeEnd SuccessActionType = "end"
	// SuccessTypeGoto branches to another step or workflow upon step success.
	SuccessTypeGoto SuccessActionType = "goto"
)

// FailureActionType represents the control flow primitive of a FailureAction.
type FailureActionType string

const (
	// FailureTypeEnd terminates the workflow with an error upon step failure.
	FailureTypeEnd FailureActionType = "end"
	// FailureTypeRetry retries the failed step or target step/workflow after an optional delay.
	FailureTypeRetry FailureActionType = "retry"
	// FailureTypeGoto branches to a fallback step or workflow upon step failure.
	FailureTypeGoto FailureActionType = "goto"
)

// Arazzo is the root object of the Arazzo Specification v1.1.0 document.
type Arazzo struct {
	// Arazzo is the semantic version number of the Arazzo Specification (e.g., "1.1.0").
	Arazzo string `json:"arazzo"`
	// Self provides the self-assigned canonical base URI of this document.
	Self string `json:"$self,omitempty"`
	// Info provides metadata about the Arazzo workflows.
	Info Info `json:"info"`
	// SourceDescriptions is the list of source descriptions (such as OpenAPI or AsyncAPI documents) linked to this Arazzo document.
	SourceDescriptions []SourceDescription `json:"sourceDescriptions"`
	// Workflows is the list of executable workflows defined in the Arazzo document.
	Workflows []Workflow `json:"workflows"`
	// Components holds reusable objects for different aspects of the Arazzo Specification.
	Components *Components `json:"components,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// Info provides metadata about the Arazzo document and its workflows.
type Info struct {
	// Title is a human-readable title of the Arazzo document.
	Title string `json:"title"`
	// Summary is a short summary of the Arazzo document.
	Summary string `json:"summary,omitempty"`
	// Description is a verbose description of the purpose of the Arazzo workflows.
	Description string `json:"description,omitempty"`
	// Version is the version string of the Arazzo document.
	Version string `json:"version"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// SourceDescription describes a source specification (e.g., OpenAPI, AsyncAPI, or Arazzo) that workflows reference.
type SourceDescription struct {
	// Name is the unique identifier for the source description.
	Name string `json:"name"`
	// URL is the URI or relative path to the source description document.
	URL string `json:"url"`
	// Type is the protocol type of the source description ("openapi", "asyncapi", or "arazzo").
	Type SourceDescriptionType `json:"type,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// Workflow describes an executable sequence or DAG of steps across API operations.
type Workflow struct {
	// WorkflowID is a unique string used to identify the workflow.
	WorkflowID string `json:"workflowId"`
	// Summary is a short summary of what the workflow does.
	Summary string `json:"summary,omitempty"`
	// Description is a verbose explanation of the workflow behavior.
	Description string `json:"description,omitempty"`
	// Inputs is a JSON Schema 2020-12 object defining the input parameters to the workflow.
	Inputs *openapi.Schema `json:"inputs,omitempty"`
	// DependsOn is a list of workflowIds that must be completed before this workflow can be processed.
	DependsOn []string `json:"dependsOn,omitempty"`
	// Steps is an ordered list of steps where each step represents a call to an API operation or another workflow.
	Steps []Step `json:"steps"`
	// SuccessActions is a list of success actions that are applicable for all steps described under this workflow.
	SuccessActions []SuccessActionOrReusable `json:"successActions,omitempty"`
	// FailureActions is a list of failure actions that are applicable for all steps described under this workflow.
	FailureActions []FailureActionOrReusable `json:"failureActions,omitempty"`
	// Outputs is a map between a friendly name and a dynamic output value or selector produced by the workflow.
	Outputs map[string]OutputValue `json:"outputs,omitempty"`
	// Parameters is a list of parameters that are applicable for all steps described under this workflow.
	Parameters []ParameterOrReusable `json:"parameters,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// Step describes a single workflow step invoking an operation, channel, or sub-workflow.
type Step struct {
	// StepID is a unique string to represent the step within the workflow.
	StepID string `json:"stepId"`
	// Description is a description of the step.
	Description string `json:"description,omitempty"`
	// OperationID is the name of an existing, resolvable operation, as defined with a unique operationId.
	OperationID string `json:"operationId,omitempty"`
	// OperationPath is a reference to a Source combined with a JSON Pointer to the operation.
	OperationPath string `json:"operationPath,omitempty"`
	// ChannelPath is a reference to an AsyncAPI Source combined with a JSON Pointer to the channel (Arazzo 1.1.0).
	ChannelPath string `json:"channelPath,omitempty"`
	// WorkflowID is the workflowId referencing another workflow within the Arazzo description or an external Arazzo source.
	WorkflowID string `json:"workflowId,omitempty"`
	// Parameters is a list of parameters that must be passed to an operation or workflow as referenced by this step.
	Parameters []ParameterOrReusable `json:"parameters,omitempty"`
	// RequestBody is the request body to pass to an operation as referenced by this step.
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	// SuccessCriteria is a list of assertions to determine the success of the step.
	SuccessCriteria []Criterion `json:"successCriteria,omitempty"`
	// OnSuccess is an array of success action objects that specify what to do upon step success.
	OnSuccess []SuccessActionOrReusable `json:"onSuccess,omitempty"`
	// OnFailure is an array of failure action objects that specify what to do upon step failure.
	OnFailure []FailureActionOrReusable `json:"onFailure,omitempty"`
	// Outputs is a map between a friendly name and a dynamic output value or selector extracted from the step result.
	Outputs map[string]OutputValue `json:"outputs,omitempty"`
	// Timeout is the maximum duration in milliseconds to wait for the step to complete (Arazzo 1.1.0).
	Timeout *int64 `json:"timeout,omitempty"`
	// CorrelationID is a runtime expression identifying the correlation token for AsyncAPI messages (Arazzo 1.1.0).
	CorrelationID string `json:"correlationId,omitempty"`
	// Action specifies the message flow intent ("send" or "receive") for AsyncAPI steps (Arazzo 1.1.0).
	Action StepActionType `json:"action,omitempty"`
	// DependsOn is a list of stepIds within the current workflow that must complete before this step executes (Arazzo 1.1.0).
	DependsOn []string `json:"dependsOn,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// ParameterValue represents a strongly-typed union value for an Arazzo Parameter or PayloadReplacement.
type ParameterValue struct {
	// StringVal holds a string literal or runtime expression when the value is a string.
	StringVal *string
	// IntVal holds an integer value when the value is an integer.
	IntVal *int64
	// FloatVal holds a floating-point value when the value is a float.
	FloatVal *float64
	// BoolVal holds a boolean value when the value is a boolean.
	BoolVal *bool
	// RawJSON holds structured JSON (object or array) when the value is a composite literal.
	RawJSON json.RawMessage
}

// NewStringValue creates a ParameterValue from a string.
// Parameter s is the string literal or runtime expression, and it returns the constructed ParameterValue.
func NewStringValue(s string) ParameterValue {
	return ParameterValue{StringVal: &s}
}

// NewIntValue creates a ParameterValue from an int64.
// Parameter i is the integer value, and it returns the constructed ParameterValue.
func NewIntValue(i int64) ParameterValue {
	return ParameterValue{IntVal: &i}
}

// NewFloatValue creates a ParameterValue from a float64.
// Parameter f is the float64 value, and it returns the constructed ParameterValue.
func NewFloatValue(f float64) ParameterValue {
	return ParameterValue{FloatVal: &f}
}

// NewBoolValue creates a ParameterValue from a bool.
// Parameter b is the boolean value, and it returns the constructed ParameterValue.
func NewBoolValue(b bool) ParameterValue {
	return ParameterValue{BoolVal: &b}
}

// String returns a string representation of the ParameterValue for code generation and expression evaluation.
func (pv ParameterValue) String() string {
	if pv.StringVal != nil {
		return *pv.StringVal
	}
	if pv.IntVal != nil {
		return strconv.FormatInt(*pv.IntVal, 10)
	}
	if pv.FloatVal != nil {
		return strconv.FormatFloat(*pv.FloatVal, 'f', -1, 64)
	}
	if pv.BoolVal != nil {
		return strconv.FormatBool(*pv.BoolVal)
	}
	if len(pv.RawJSON) > 0 {
		return string(pv.RawJSON)
	}
	return ""
}

// MarshalJSON serializes ParameterValue into its underlying JSON scalar or composite value.
func (pv ParameterValue) MarshalJSON() ([]byte, error) {
	if pv.StringVal != nil {
		return json.Marshal(*pv.StringVal)
	}
	if pv.IntVal != nil {
		return json.Marshal(*pv.IntVal)
	}
	if pv.FloatVal != nil {
		return json.Marshal(*pv.FloatVal)
	}
	if pv.BoolVal != nil {
		return json.Marshal(*pv.BoolVal)
	}
	if len(pv.RawJSON) > 0 {
		return pv.RawJSON, nil
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes a JSON scalar or composite value into ParameterValue without using any/interface{}.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (pv *ParameterValue) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*pv = ParameterValue{}
		return nil
	}
	switch trimmed[0] {
	case '"':
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		*pv = ParameterValue{StringVal: &s}
		return nil
	case 't', 'f':
		var b bool
		if err := json.Unmarshal(trimmed, &b); err != nil {
			return err
		}
		*pv = ParameterValue{BoolVal: &b}
		return nil
	case '{', '[':
		if !json.Valid(trimmed) {
			return fmt.Errorf("invalid composite JSON in ParameterValue")
		}
		rawCopy := make(json.RawMessage, len(trimmed))
		copy(rawCopy, trimmed)
		*pv = ParameterValue{RawJSON: rawCopy}
		return nil
	default:
		if !bytes.ContainsAny(trimmed, ".eE") {
			var i int64
			if err := json.Unmarshal(trimmed, &i); err == nil {
				*pv = ParameterValue{IntVal: &i}
				return nil
			}
		}
		var f float64
		if err := json.Unmarshal(trimmed, &f); err != nil {
			return err
		}
		*pv = ParameterValue{FloatVal: &f}
		return nil
	}
}

// Parameter describes a single step or workflow parameter binding.
type Parameter struct {
	// Name is the name of the parameter.
	Name string `json:"name"`
	// In is the named location of the parameter ("path", "query", "querystring", "header", or "cookie").
	In ParameterIn `json:"in,omitempty"`
	// Value is the constant value or runtime expression passed for the parameter.
	Value ParameterValue `json:"value"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// ReusableObject represents a reference to a reusable component in Components.
type ReusableObject struct {
	// Reference is a runtime expression used to reference the desired object (e.g. "$components.parameters.page").
	Reference string `json:"reference"`
	// Value is an optional override value for a referenced parameter object.
	Value *ParameterValue `json:"value,omitempty"`
}

// ParameterOrReusable is a strongly-typed discriminated union holding either a Parameter or a ReusableObject.
type ParameterOrReusable struct {
	// Parameter is the inline Parameter definition when not a reference.
	Parameter *Parameter
	// Reusable is the component reference when referencing $components.parameters.
	Reusable *ReusableObject
}

// MarshalJSON serializes ParameterOrReusable as either a ReusableObject or a Parameter.
func (pr ParameterOrReusable) MarshalJSON() ([]byte, error) {
	if pr.Reusable != nil {
		return json.Marshal(pr.Reusable)
	}
	if pr.Parameter != nil {
		return json.Marshal(pr.Parameter)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes JSON into either a ReusableObject (if "reference" is present) or a Parameter.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (pr *ParameterOrReusable) UnmarshalJSON(data []byte) error {
	var probe struct {
		// Reference checks for the presence of a reusable reference field.
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe.Reference != "" {
		var r ReusableObject
		_ = json.Unmarshal(data, &r)
		pr.Reusable = &r
		pr.Parameter = nil
		return nil
	}
	var p Parameter
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	pr.Parameter = &p
	pr.Reusable = nil
	return nil
}

// SuccessAction describes a single action to take upon step or workflow success.
type SuccessAction struct {
	// Name is the case-sensitive name of the success action.
	Name string `json:"name"`
	// Type is the type of action to take ("end" or "goto").
	Type SuccessActionType `json:"type"`
	// WorkflowID is the workflowId to transfer to when Type is "goto".
	WorkflowID string `json:"workflowId,omitempty"`
	// StepID is the stepId to transfer to when Type is "goto".
	StepID string `json:"stepId,omitempty"`
	// Parameters is a list of parameters to pass to a target workflow or step.
	Parameters []ParameterOrReusable `json:"parameters,omitempty"`
	// Criteria is a list of assertions to determine if this action SHALL be executed.
	Criteria []Criterion `json:"criteria,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// SuccessActionOrReusable is a strongly-typed discriminated union holding either a SuccessAction or a ReusableObject.
type SuccessActionOrReusable struct {
	// Action is the inline SuccessAction definition when not a reference.
	Action *SuccessAction
	// Reusable is the component reference when referencing $components.successActions.
	Reusable *ReusableObject
}

// MarshalJSON serializes SuccessActionOrReusable as either a ReusableObject or a SuccessAction.
func (sr SuccessActionOrReusable) MarshalJSON() ([]byte, error) {
	if sr.Reusable != nil {
		return json.Marshal(sr.Reusable)
	}
	if sr.Action != nil {
		return json.Marshal(sr.Action)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes JSON into either a ReusableObject or a SuccessAction.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (sr *SuccessActionOrReusable) UnmarshalJSON(data []byte) error {
	var probe struct {
		// Reference checks for the presence of a reusable reference field.
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe.Reference != "" {
		var r ReusableObject
		_ = json.Unmarshal(data, &r)
		sr.Reusable = &r
		sr.Action = nil
		return nil
	}
	var a SuccessAction
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	sr.Action = &a
	sr.Reusable = nil
	return nil
}

// FailureAction describes a single action to take upon step or workflow failure.
type FailureAction struct {
	// Name is the case-sensitive name of the failure action.
	Name string `json:"name"`
	// Type is the type of action to take ("end", "retry", or "goto").
	Type FailureActionType `json:"type"`
	// WorkflowID is the workflowId to transfer to when Type is "goto" or "retry".
	WorkflowID string `json:"workflowId,omitempty"`
	// StepID is the stepId to transfer to when Type is "goto" or "retry".
	StepID string `json:"stepId,omitempty"`
	// Parameters is a list of parameters to pass to a target workflow or step.
	Parameters []ParameterOrReusable `json:"parameters,omitempty"`
	// RetryAfter is a non-negative decimal indicating the seconds to delay after step failure before retrying.
	RetryAfter *float64 `json:"retryAfter,omitempty"`
	// RetryLimit is a non-negative integer indicating how many attempts to retry the step may be attempted before failing.
	RetryLimit *int `json:"retryLimit,omitempty"`
	// Criteria is a list of assertions to determine if this action SHALL be executed.
	Criteria []Criterion `json:"criteria,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// FailureActionOrReusable is a strongly-typed discriminated union holding either a FailureAction or a ReusableObject.
type FailureActionOrReusable struct {
	// Action is the inline FailureAction definition when not a reference.
	Action *FailureAction
	// Reusable is the component reference when referencing $components.failureActions.
	Reusable *ReusableObject
}

// MarshalJSON serializes FailureActionOrReusable as either a ReusableObject or a FailureAction.
func (fr FailureActionOrReusable) MarshalJSON() ([]byte, error) {
	if fr.Reusable != nil {
		return json.Marshal(fr.Reusable)
	}
	if fr.Action != nil {
		return json.Marshal(fr.Action)
	}
	return []byte("null"), nil
}

// UnmarshalJSON deserializes JSON into either a ReusableObject or a FailureAction.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (fr *FailureActionOrReusable) UnmarshalJSON(data []byte) error {
	var probe struct {
		// Reference checks for the presence of a reusable reference field.
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if probe.Reference != "" {
		var r ReusableObject
		_ = json.Unmarshal(data, &r)
		fr.Reusable = &r
		fr.Action = nil
		return nil
	}
	var a FailureAction
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	fr.Action = &a
	fr.Reusable = nil
	return nil
}

// CriterionExpressionType specifies the dialect type and version of an expression in a Criterion, Selector, or PayloadReplacement.
type CriterionExpressionType struct {
	// Type is the type of condition or selector to be applied ("jsonpath", "xpath", or "jsonpointer").
	Type string `json:"type"`
	// Version is the short hand string representing the version of the expression type (e.g., "rfc9535", "xpath-30", "rfc6901").
	Version string `json:"version"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// CriterionType is a strongly-typed union representing either a simple string dialect ("simple", "regex", "jsonpath", "xpath") or a CriterionExpressionType object.
type CriterionType struct {
	// SimpleType holds the string literal when the criterion type is specified as a string.
	SimpleType string
	// ExpressionType holds the structured CriterionExpressionType when specified as an object.
	ExpressionType *CriterionExpressionType
}

// EffectiveType returns the normalized dialect string ("simple", "regex", "jsonpath", "xpath", or "jsonpointer").
func (ct *CriterionType) EffectiveType() string {
	if ct == nil {
		return "simple"
	}
	if ct.ExpressionType != nil && ct.ExpressionType.Type != "" {
		return ct.ExpressionType.Type
	}
	if ct.SimpleType != "" {
		return ct.SimpleType
	}
	return "simple"
}

// MarshalJSON serializes CriterionType as either a JSON string or a CriterionExpressionType JSON object.
func (ct CriterionType) MarshalJSON() ([]byte, error) {
	if ct.ExpressionType != nil {
		return json.Marshal(ct.ExpressionType)
	}
	return json.Marshal(ct.SimpleType)
}

// UnmarshalJSON deserializes CriterionType from either a JSON string or a CriterionExpressionType object.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (ct *CriterionType) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*ct = CriterionType{}
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		ct.SimpleType = s
		ct.ExpressionType = nil
		return nil
	}
	var expr CriterionExpressionType
	if err := json.Unmarshal(trimmed, &expr); err != nil {
		return err
	}
	ct.ExpressionType = &expr
	ct.SimpleType = ""
	return nil
}

// Criterion describes an assertion context, condition expression, and evaluation dialect.
type Criterion struct {
	// Context is a runtime expression used to set the context for the condition to be applied on.
	Context string `json:"context,omitempty"`
	// Condition is the condition to apply (e.g., "$statusCode == 200").
	Condition string `json:"condition"`
	// Type specifies the condition dialect ("simple", "regex", "jsonpath", "xpath", or an ExpressionType object).
	Type *CriterionType `json:"type,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// Selector describes a structured query for extracting dynamic output variables in Arazzo 1.1.0.
type Selector struct {
	// Context is a runtime expression defining the source data context (e.g., "$response.body").
	Context string `json:"context"`
	// Selector is the query expression string (e.g., "$.items[0].id" or "/items/0/id").
	Selector string `json:"selector"`
	// Type specifies the selector query dialect ("jsonpath", "xpath", "jsonpointer", or an ExpressionType object).
	Type *CriterionType `json:"type,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// OutputValue is a strongly-typed discriminated union representing either a runtime expression string or an Arazzo 1.1.0 Selector object.
type OutputValue struct {
	// Expression holds the runtime expression string when the output is defined as a string.
	Expression string
	// Selector holds the structured Selector object when the output is defined as a Selector.
	Selector *Selector
}

// MarshalJSON serializes OutputValue as either a JSON string or a Selector object.
func (ov OutputValue) MarshalJSON() ([]byte, error) {
	if ov.Selector != nil {
		return json.Marshal(ov.Selector)
	}
	return json.Marshal(ov.Expression)
}

// UnmarshalJSON deserializes OutputValue from either a JSON string or a Selector object.
// Parameter data is the raw JSON byte slice, and it returns any unmarshaling error.
func (ov *OutputValue) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		*ov = OutputValue{}
		return nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		ov.Expression = s
		ov.Selector = nil
		return nil
	}
	var sel Selector
	if err := json.Unmarshal(trimmed, &sel); err != nil {
		return err
	}
	ov.Selector = &sel
	ov.Expression = ""
	return nil
}

// RequestBody describes the request payload and dynamic replacements for a Step.
type RequestBody struct {
	// ContentType is the Content-Type for the request content.
	ContentType string `json:"contentType,omitempty"`
	// Payload is a raw JSON value, string template, or expression representing the request body payload.
	Payload json.RawMessage `json:"payload,omitempty"`
	// Replacements is a list of locations and values to replace within Payload before transmission.
	Replacements []PayloadReplacement `json:"replacements,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// PayloadReplacement describes a targeted field replacement within a RequestBody Payload.
type PayloadReplacement struct {
	// Target is a JSON Pointer, JSONPath, or XPath expression to the location within Payload to be replaced.
	Target string `json:"target"`
	// TargetSelectorType specifies the selector dialect for Target in Arazzo 1.1.0.
	TargetSelectorType *CriterionType `json:"targetSelectorType,omitempty"`
	// Value is the literal value or runtime expression to set at the target location.
	Value ParameterValue `json:"value"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}

// Components holds a set of reusable objects for different aspects of the Arazzo Specification.
type Components struct {
	// Inputs holds reusable JSON Schema 2020-12 objects for workflow inputs.
	Inputs map[string]openapi.Schema `json:"inputs,omitempty"`
	// Parameters holds reusable Parameter objects.
	Parameters map[string]Parameter `json:"parameters,omitempty"`
	// SuccessActions holds reusable SuccessAction objects.
	SuccessActions map[string]SuccessAction `json:"successActions,omitempty"`
	// FailureActions holds reusable FailureAction objects.
	FailureActions map[string]FailureAction `json:"failureActions,omitempty"`
	// Extensions holds specification extensions (^x-).
	Extensions map[string]json.RawMessage `json:"-"`
}
