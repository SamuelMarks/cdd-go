package responses

import (
	"github.com/dave/dst"
	"testing"
)

func TestParseResponses(t *testing.T) {
	results := &dst.FieldList{
		List: []*dst.Field{
			{Type: &dst.StarExpr{X: dst.NewIdent("User")}},
			{
				Names: []*dst.Ident{dst.NewIdent("X-RateLimit")},
				Type:  dst.NewIdent("int"),
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Start: dst.Decorations{
							"// Rate limit",
							"// Required: true",
						},
					},
				},
			},
			{Type: dst.NewIdent("error")},
		},
	}

	resps := Parse(results)
	if resps == nil {
		t.Fatalf("expected resps")
	}
	if resp, ok := resps["200"]; !ok {
		t.Errorf("expected 200")
	} else {
		if mt, ok := resp.Content["application/json"]; !ok || mt.Schema.Ref != "#/components/schemas/User" {
			t.Errorf("expected User ref")
		}
		if resp.Headers == nil || len(resp.Headers) != 1 {
			t.Errorf("expected headers")
		} else if h, ok := resp.Headers["X-RateLimit"]; !ok || h.Description != "Rate limit" || !h.Required {
			t.Errorf("expected parsed header")
		}
	}
}

func TestParseResponsesHeadersOnly(t *testing.T) {
	results := &dst.FieldList{
		List: []*dst.Field{
			{
				Names: []*dst.Ident{dst.NewIdent("X-Test")},
				Type:  dst.NewIdent("string"),
			},
			{Type: dst.NewIdent("error")},
		},
	}

	resps := Parse(results)
	if resp, ok := resps["200"]; !ok {
		t.Errorf("expected 200")
	} else if len(resp.Headers) != 1 {
		t.Errorf("expected headers")
	}
}

func TestParseResponsesNonStar(t *testing.T) {
	results := &dst.FieldList{
		List: []*dst.Field{
			{Type: dst.NewIdent("User")},
		},
	}

	resps := Parse(results)
	if resps == nil {
		t.Fatalf("expected resps")
	}
	if resp, ok := resps["200"]; !ok {
		t.Errorf("expected 200")
	} else if mt, ok := resp.Content["application/json"]; !ok || mt.Schema.Ref != "#/components/schemas/User" {
		t.Errorf("expected User ref")
	}
}

func TestParseResponsesDefault(t *testing.T) {
	results := &dst.FieldList{
		List: []*dst.Field{
			{Type: dst.NewIdent("error")},
		},
	}
	resps := Parse(results)
	if resp, ok := resps["default"]; !ok {
		t.Errorf("expected default")
	} else if resp.Description != "Default response" {
		t.Errorf("expected description")
	}
}

func TestParseResponsesNil(t *testing.T) {
	if Parse(nil) != nil {
		t.Errorf("expected nil")
	}
}
