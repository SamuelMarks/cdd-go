package mcp

// QueryASTResource handles querying the AST as an MCP Resource.
func QueryASTResource(uri string) (*Resource, error) {
	return &Resource{
		Name:        "AST Query",
		URI:         uri,
		Description: "Queries internal AST structures as MCP resources",
		MimeType:    "application/json",
	}, nil
}

// InMemoryGenerationRouter natively bindings to run the generator core directly.
func InMemoryGenerationRouter(input map[string]interface{}) (interface{}, error) {
	// Stub: Natively routes requests to cdd.GenerateFromOpenApi or cdd.GenerateToOpenApi in memory.
	return map[string]interface{}{"status": "generated in memory"}, nil
}
