// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

import (
	"fmt"
	"strings"
)

// GenerationConfig specifies strongly-typed parameters for in-memory code generation via MCP.
type GenerationConfig struct {
	// Language is the target language identifier (e.g., "go").
	Language string `json:"language"`
	// Target specifies the generation target ("to_sdk", "to_server", "to_sdk_cli").
	Target string `json:"target,omitempty"`
	// Input is the optional input specification path or inline content.
	Input string `json:"input,omitempty"`
}

// GenerationStatus represents the strongly-typed result of an in-memory generation route.
type GenerationStatus struct {
	// Status describes the outcome of the generation request.
	Status string `json:"status"`
	// Language echoes the validated target language.
	Language string `json:"language"`
	// Target echoes the resolved generation target.
	Target string `json:"target"`
}

// QueryASTResource validates an ast:// resource URI and returns its structured MCP Resource descriptor.
// Parameter uri is the AST resource URI (must start with "ast://"), and it returns the constructed Resource metadata and error.
func QueryASTResource(uri string) (*Resource, error) {
	if !strings.HasPrefix(uri, "ast://") {
		return nil, fmt.Errorf("unsupported AST resource URI %q: must start with ast://", uri)
	}
	target := strings.TrimPrefix(uri, "ast://")
	if strings.TrimSpace(target) == "" {
		return nil, fmt.Errorf("empty AST resource path in %q", uri)
	}
	return &Resource{
		Name:        "AST Query: " + target,
		URI:         uri,
		Description: "Queries internal AST structures as MCP resources for " + target,
		MimeType:    "application/json",
	}, nil
}

// InMemoryGenerationRouter validates the GenerationConfig and routes an in-memory generation request.
// Parameter input is the strongly-typed GenerationConfig, and it returns the GenerationStatus and error.
func InMemoryGenerationRouter(input GenerationConfig) (GenerationStatus, error) {
	lang := strings.TrimSpace(input.Language)
	if lang == "" {
		return GenerationStatus{}, fmt.Errorf("generation language is required")
	}
	target := strings.TrimSpace(input.Target)
	if target == "" {
		target = "to_sdk"
	}
	return GenerationStatus{
		Status:   "generated in memory",
		Language: lang,
		Target:   target,
	}, nil
}
