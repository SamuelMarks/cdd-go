// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import (
	"fmt"
	"strings"
)

// ExprRootKind represents the canonical Arazzo 1.1.0 runtime expression root namespace.
type ExprRootKind string

const (
	// ExprLiteral indicates a constant string without a runtime expression root.
	ExprLiteral ExprRootKind = "literal"
	// ExprURL indicates the $url runtime expression.
	ExprURL ExprRootKind = "$url"
	// ExprMethod indicates the $method runtime expression.
	ExprMethod ExprRootKind = "$method"
	// ExprStatusCode indicates the $statusCode runtime expression.
	ExprStatusCode ExprRootKind = "$statusCode"
	// ExprRequestHeader indicates the $request.header.<name> runtime expression.
	ExprRequestHeader ExprRootKind = "$request.header"
	// ExprRequestQuery indicates the $request.query.<name> runtime expression.
	ExprRequestQuery ExprRootKind = "$request.query"
	// ExprRequestPath indicates the $request.path.<name> runtime expression.
	ExprRequestPath ExprRootKind = "$request.path"
	// ExprRequestBody indicates the $request.body runtime expression.
	ExprRequestBody ExprRootKind = "$request.body"
	// ExprResponseHeader indicates the $response.header.<name> runtime expression.
	ExprResponseHeader ExprRootKind = "$response.header"
	// ExprResponseBody indicates the $response.body runtime expression.
	ExprResponseBody ExprRootKind = "$response.body"
	// ExprMessageHeader indicates the $message.header.<name> AsyncAPI runtime expression (Arazzo 1.1.0).
	ExprMessageHeader ExprRootKind = "$message.header"
	// ExprMessagePayload indicates the $message.payload AsyncAPI runtime expression (Arazzo 1.1.0).
	ExprMessagePayload ExprRootKind = "$message.payload"
	// ExprInputs indicates the $inputs.<name> runtime expression.
	ExprInputs ExprRootKind = "$inputs"
	// ExprOutputs indicates the $outputs.<name> runtime expression.
	ExprOutputs ExprRootKind = "$outputs"
	// ExprSteps indicates the $steps.<stepId>.* runtime expression.
	ExprSteps ExprRootKind = "$steps"
	// ExprWorkflows indicates the $workflows.<workflowId>.* runtime expression.
	ExprWorkflows ExprRootKind = "$workflows"
	// ExprSourceDescriptions indicates the $sourceDescriptions.<name>.* runtime expression.
	ExprSourceDescriptions ExprRootKind = "$sourceDescriptions"
	// ExprComponents indicates the $components.<type>.<name> runtime expression.
	ExprComponents ExprRootKind = "$components"
	// ExprSelf indicates the $self runtime expression (Arazzo 1.1.0).
	ExprSelf ExprRootKind = "$self"
)

// ParsedExpression represents a parsed Arazzo 1.1.0 runtime expression or template.
type ParsedExpression struct {
	// Raw is the original expression string.
	Raw string
	// Kind is the classified root namespace of the runtime expression.
	Kind ExprRootKind
	// TargetID is the primary identifier (such as header/query/path name, input/output key, stepId, workflowId, source name, or component type).
	TargetID string
	// SubKey is the secondary identifier (such as output name on $steps.<stepId>.outputs.<SubKey> or component name on $components.<type>.<SubKey>).
	SubKey string
	// Pointer is the optional JSON Pointer or property path suffix (e.g., "#/data/id" or "/id").
	Pointer string
	// IsTemplate indicates whether the expression is embedded inside a string template with curly braces (e.g., "Bearer {$inputs.token}").
	IsTemplate bool
}

// ParseRuntimeExpression parses and classifies an Arazzo 1.1.0 runtime expression or string template.
// Parameter expr is the raw expression string, and it returns the classified ParsedExpression and any validation error.
func ParseRuntimeExpression(expr string) (ParsedExpression, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return ParsedExpression{Raw: expr, Kind: ExprLiteral}, nil
	}

	// Check for embedded template braces like "Bearer {$steps.login.outputs.token}"
	if !strings.HasPrefix(trimmed, "$") {
		startIdx := strings.Index(trimmed, "{$")
		if startIdx >= 0 {
			endIdx := strings.Index(trimmed[startIdx:], "}")
			if endIdx < 0 {
				return ParsedExpression{}, fmt.Errorf("unclosed embedded runtime expression in %q", expr)
			}
			inner := trimmed[startIdx+1 : startIdx+endIdx]
			innerParsed, err := ParseRuntimeExpression(inner)
			if err != nil {
				return ParsedExpression{}, err
			}
			innerParsed.Raw = expr
			innerParsed.IsTemplate = true
			return innerParsed, nil
		}
		return ParsedExpression{Raw: expr, Kind: ExprLiteral}, nil
	}

	var pointer string
	base := trimmed
	if hashIdx := strings.Index(base, "#/"); hashIdx >= 0 {
		pointer = base[hashIdx:]
		base = base[:hashIdx]
	}

	switch {
	case base == "$url":
		return ParsedExpression{Raw: expr, Kind: ExprURL, Pointer: pointer}, nil
	case base == "$method":
		return ParsedExpression{Raw: expr, Kind: ExprMethod, Pointer: pointer}, nil
	case base == "$statusCode":
		return ParsedExpression{Raw: expr, Kind: ExprStatusCode, Pointer: pointer}, nil
	case base == "$self":
		return ParsedExpression{Raw: expr, Kind: ExprSelf, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$request.header."):
		name := strings.TrimPrefix(base, "$request.header.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty header name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprRequestHeader, TargetID: name, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$request.query."):
		name := strings.TrimPrefix(base, "$request.query.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty query name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprRequestQuery, TargetID: name, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$request.path."):
		name := strings.TrimPrefix(base, "$request.path.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty path parameter name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprRequestPath, TargetID: name, Pointer: pointer}, nil
	case base == "$request.body" || strings.HasPrefix(base, "$request.body."):
		prop := strings.TrimPrefix(base, "$request.body")
		prop = strings.TrimPrefix(prop, ".")
		return ParsedExpression{Raw: expr, Kind: ExprRequestBody, TargetID: prop, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$response.header."):
		name := strings.TrimPrefix(base, "$response.header.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty response header name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprResponseHeader, TargetID: name, Pointer: pointer}, nil
	case base == "$response.body" || strings.HasPrefix(base, "$response.body."):
		prop := strings.TrimPrefix(base, "$response.body")
		prop = strings.TrimPrefix(prop, ".")
		return ParsedExpression{Raw: expr, Kind: ExprResponseBody, TargetID: prop, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$message.header."):
		name := strings.TrimPrefix(base, "$message.header.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty message header name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprMessageHeader, TargetID: name, Pointer: pointer}, nil
	case base == "$message.payload" || strings.HasPrefix(base, "$message.payload."):
		prop := strings.TrimPrefix(base, "$message.payload")
		prop = strings.TrimPrefix(prop, ".")
		return ParsedExpression{Raw: expr, Kind: ExprMessagePayload, TargetID: prop, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$inputs."):
		name := strings.TrimPrefix(base, "$inputs.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty input name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprInputs, TargetID: name, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$outputs."):
		name := strings.TrimPrefix(base, "$outputs.")
		if name == "" {
			return ParsedExpression{}, fmt.Errorf("empty output name in %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprOutputs, TargetID: name, Pointer: pointer}, nil
	case strings.HasPrefix(base, "$steps."):
		rest := strings.TrimPrefix(base, "$steps.")
		parts := strings.SplitN(rest, ".", 3)
		if len(parts) < 3 || parts[0] == "" || parts[1] != "outputs" || parts[2] == "" {
			return ParsedExpression{}, fmt.Errorf("invalid $steps expression %q: expected $steps.<stepId>.outputs.<name>", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprSteps, TargetID: parts[0], SubKey: parts[2], Pointer: pointer}, nil
	case strings.HasPrefix(base, "$workflows."):
		rest := strings.TrimPrefix(base, "$workflows.")
		parts := strings.SplitN(rest, ".", 4)
		if len(parts) < 3 || parts[0] == "" || parts[2] == "" {
			return ParsedExpression{}, fmt.Errorf("invalid $workflows expression %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprWorkflows, TargetID: parts[0], SubKey: strings.Join(parts[2:], "."), Pointer: pointer}, nil
	case strings.HasPrefix(base, "$sourceDescriptions."):
		rest := strings.TrimPrefix(base, "$sourceDescriptions.")
		parts := strings.SplitN(rest, ".", 2)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return ParsedExpression{}, fmt.Errorf("invalid $sourceDescriptions expression %q", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprSourceDescriptions, TargetID: parts[0], SubKey: parts[1], Pointer: pointer}, nil
	case strings.HasPrefix(base, "$components."):
		rest := strings.TrimPrefix(base, "$components.")
		parts := strings.SplitN(rest, ".", 2)
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			return ParsedExpression{}, fmt.Errorf("invalid $components expression %q: expected $components.<type>.<name>", expr)
		}
		return ParsedExpression{Raw: expr, Kind: ExprComponents, TargetID: parts[0], SubKey: parts[1], Pointer: pointer}, nil
	default:
		return ParsedExpression{}, fmt.Errorf("unrecognized runtime expression root in %q", expr)
	}
}
