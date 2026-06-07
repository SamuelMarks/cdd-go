package mcp

import (
	"testing"
)

func TestMetaGenerator(t *testing.T) {
	r, err := QueryASTResource("ast://model/User")
	if err != nil {
		t.Errorf("expected no error")
	}
	if r.Name != "AST Query" {
		t.Errorf("expected AST Query name")
	}

	res, err := InMemoryGenerationRouter(map[string]interface{}{"lang": "go"})
	if err != nil {
		t.Errorf("expected no error")
	}
	if res.(map[string]interface{})["status"] != "generated in memory" {
		t.Errorf("expected generated status")
	}
}
