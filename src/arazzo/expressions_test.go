// Package arazzo provides strongly-typed structures, parsers, validators, and emitters for the Arazzo Specification v1.1.0.
package arazzo

import "testing"

func TestParseRuntimeExpression(t *testing.T) {
	validCases := []struct {
		expr string
		kind ExprRootKind
		id   string
		sub  string
		ptr  string
		tmpl bool
	}{
		{"", ExprLiteral, "", "", "", false},
		{"plain_literal", ExprLiteral, "", "", "", false},
		{"$url", ExprURL, "", "", "", false},
		{"$method", ExprMethod, "", "", "", false},
		{"$statusCode", ExprStatusCode, "", "", "", false},
		{"$self", ExprSelf, "", "", "", false},
		{"$request.header.Authorization", ExprRequestHeader, "Authorization", "", "", false},
		{"$request.query.page", ExprRequestQuery, "page", "", "", false},
		{"$request.path.petId", ExprRequestPath, "petId", "", "", false},
		{"$request.body#/user/id", ExprRequestBody, "", "", "#/user/id", false},
		{"$request.body.user", ExprRequestBody, "user", "", "", false},
		{"$response.header.Retry-After", ExprResponseHeader, "Retry-After", "", "", false},
		{"$response.body#/token", ExprResponseBody, "", "", "#/token", false},
		{"$response.body.token", ExprResponseBody, "token", "", "", false},
		{"$message.header.correlationId", ExprMessageHeader, "correlationId", "", "", false},
		{"$message.payload#/eventId", ExprMessagePayload, "", "", "#/eventId", false},
		{"$message.payload.eventId", ExprMessagePayload, "eventId", "", "", false},
		{"$inputs.username", ExprInputs, "username", "", "", false},
		{"$outputs.resultId", ExprOutputs, "resultId", "", "", false},
		{"$steps.loginStep.outputs.token#/jwt", ExprSteps, "loginStep", "token", "#/jwt", false},
		{"$workflows.authWf.outputs.token", ExprWorkflows, "authWf", "token", "", false},
		{"$sourceDescriptions.petstore.url", ExprSourceDescriptions, "petstore", "url", "", false},
		{"$components.parameters.sharedHeader", ExprComponents, "parameters", "sharedHeader", "", false},
		{"Bearer {$steps.loginStep.outputs.token}", ExprSteps, "loginStep", "token", "", true},
	}

	for _, tc := range validCases {
		parsed, err := ParseRuntimeExpression(tc.expr)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.expr, err)
		}
		if parsed.Kind != tc.kind {
			t.Errorf("for %q expected kind %q, got %q", tc.expr, tc.kind, parsed.Kind)
		}
		if parsed.TargetID != tc.id {
			t.Errorf("for %q expected TargetID %q, got %q", tc.expr, tc.id, parsed.TargetID)
		}
		if parsed.SubKey != tc.sub {
			t.Errorf("for %q expected SubKey %q, got %q", tc.expr, tc.sub, parsed.SubKey)
		}
		if parsed.Pointer != tc.ptr {
			t.Errorf("for %q expected Pointer %q, got %q", tc.expr, tc.ptr, parsed.Pointer)
		}
		if parsed.IsTemplate != tc.tmpl {
			t.Errorf("for %q expected IsTemplate %v, got %v", tc.expr, tc.tmpl, parsed.IsTemplate)
		}
	}

	invalidCases := []string{
		"Bearer {$steps.loginStep.outputs.token",
		"Bearer {$unknown.root}",
		"$request.header.",
		"$request.query.",
		"$request.path.",
		"$response.header.",
		"$message.header.",
		"$inputs.",
		"$outputs.",
		"$steps.onlyOne",
		"$steps..outputs.x",
		"$steps.s1.notoutputs.x",
		"$workflows.onlyOne",
		"$sourceDescriptions.onlyOne",
		"$components.onlyOne",
		"$unknownRoot",
	}

	for _, invalid := range invalidCases {
		if _, err := ParseRuntimeExpression(invalid); err == nil {
			t.Errorf("expected error for invalid expression %q", invalid)
		}
	}
}
