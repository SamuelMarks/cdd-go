// Package ir provides the universal Intermediate Representation (IR) that decouples
// Swagger 2.0, OpenAPI 3.2.0, Arazzo 1.1.0, and Go AST emitters and parsers.
package ir

import (
	"github.com/SamuelMarks/cdd-go/src/arazzo"
	"github.com/SamuelMarks/cdd-go/src/openapi"
)

// Document is the canonical Intermediate Representation (IR) holding normalized
// API models, endpoints, security schemes, metadata, and orchestration workflows.
type Document struct {
	// Metadata holds API-level title, version, servers, tags, and external documentation.
	Metadata Metadata
	// Models holds normalized domain schemas keyed by model name.
	Models map[string]openapi.Schema
	// Endpoints holds normalized HTTP path items and operations keyed by URL path.
	Endpoints map[string]openapi.PathItem
	// SecuritySchemes holds normalized authentication and authorization schemes.
	SecuritySchemes map[string]openapi.SecurityScheme
	// Security holds global security requirements applied across endpoints.
	Security []openapi.SecurityRequirement
	// Components holds normalized reusable parameters, headers, request bodies, responses, examples, links, and callbacks.
	Components openapi.Components
	// Webhooks holds normalized incoming webhook definitions keyed by webhook name.
	Webhooks map[string]openapi.PathItem
	// Workflows holds optional Arazzo 1.1.0 workflows associated with the API.
	Workflows []arazzo.Workflow
}

// Metadata holds normalized API metadata across Swagger 2.0 and OpenAPI 3.2.0.
type Metadata struct {
	// OpenAPIVersion is the normalized target OpenAPI version (defaults to "3.2.0").
	OpenAPIVersion string
	// Info holds title, version, description, contact, and license metadata.
	Info openapi.Info
	// Servers holds normalized server endpoints (including converted Swagger 2.0 host/basePath/schemes).
	Servers []openapi.Server
	// Tags holds endpoint grouping tags.
	Tags []openapi.Tag
	// ExternalDocs holds top-level external documentation links.
	ExternalDocs *openapi.ExternalDocs
	// Self is the optional $self URI of the OpenAPI 3.2.0 document.
	Self string
	// JSONSchemaDialect is the optional JSON Schema dialect URI.
	JSONSchemaDialect string
}
