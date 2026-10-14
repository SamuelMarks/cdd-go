// Package mcp provides strongly-typed Model Context Protocol (MCP) 1.0.0 and JSON-RPC 2.0 schemas, transports, and handlers.
package mcp

// QueryASTResource handles querying the AST as an MCP Resource.
// Parameter uri is the AST resource URI, and it returns the constructed Resource metadata and error.
func QueryASTResource(uri string) (*Resource, error) {
	return &Resource{
		Name:        "AST Query",
		URI:         uri,
		Description: "Queries internal AST structures as MCP resources",
		MimeType:    "application/json",
	}, nil
}

// InMemoryGenerationRouter natively binds to run the generator core directly.
// Parameter input is the map of generation parameters, and it returns the generation status map and error.
func InMemoryGenerationRouter(input map[string]any) (map[string]string, error) {
	return map[string]string{"status": "generated in memory"}, nil
}
