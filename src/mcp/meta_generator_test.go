package mcp

import (
	"testing"
)

func TestMetaGenerator(t *testing.T) {
	r, err := QueryASTResource("ast://model/User")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Name != "AST Query: model/User" || r.URI != "ast://model/User" {
		t.Errorf("unexpected resource: %+v", r)
	}

	if _, err := QueryASTResource("http://invalid"); err == nil {
		t.Errorf("expected error for non-ast:// scheme")
	}
	if _, err := QueryASTResource("ast://   "); err == nil {
		t.Errorf("expected error for empty ast:// path")
	}

	res, err := InMemoryGenerationRouter(GenerationConfig{Language: "go"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Status != "generated in memory" || res.Language != "go" || res.Target != "to_sdk" {
		t.Errorf("unexpected GenerationStatus: %+v", res)
	}

	resExplicit, err := InMemoryGenerationRouter(GenerationConfig{Language: "go", Target: "to_server"})
	if err != nil || resExplicit.Target != "to_server" {
		t.Errorf("expected explicit target to_server, got %+v, %v", resExplicit, err)
	}

	if _, err := InMemoryGenerationRouter(GenerationConfig{}); err == nil {
		t.Errorf("expected error when Language is empty")
	}
}
